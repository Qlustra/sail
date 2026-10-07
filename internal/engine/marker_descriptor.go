package engine

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/capability"
	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
	"github.com/qlustra/sail/internal/tags"
)

type markerRole string

const (
	legendRole   markerRole = "legend"
	landmarkRole markerRole = "landmark"
	regionRole   markerRole = "region"
)

// A markerDescriptor records the rules of a reusable marker declaration.
// Lane and spot descriptors arrange its children; subject metadata belongs
// to populated markers in a chart.
type markerDescriptor struct {
	typ         reflect.Type
	role        markerRole
	mark        mark
	acceptance  acceptance
	qualifiers  []qualifierDescriptor
	features    []featureDescriptor
	barrels     []barrelDescriptor
	lanes       []laneDescriptor
	carrierPath []int
}

type featureDescriptor struct {
	symbol          string
	declarationPath []int
}

// newMarkerDescriptor starts a local description; the explorer records it only after
// its fields, mark, and acceptance have passed validation.
func newMarkerDescriptor(t reflect.Type, role markerRole) (*markerDescriptor, error) {
	path, err := reflectutils.EmbeddedPath(t, role.carrier())
	if err != nil {
		return nil, err
	}
	return &markerDescriptor{typ: t, role: role, carrierPath: path}, nil
}

func (role markerRole) carrier() reflect.Type {
	switch role {
	case legendRole:
		return reflect.TypeFor[decl.Legend]()
	case regionRole:
		return reflect.TypeFor[decl.Region]()
	case landmarkRole:
		return reflect.TypeFor[decl.Landmark]()
	}
	return nil
}

func (d *markerDescriptor) collectMark(tag string, surrogate any) error {
	var family string
	var err error
	// The explorer checks required providers before allocating the surrogate.
	if d.role == legendRole {
		provider := surrogate.(decl.MarkProvider)
		tag, err = capability.InvokeDeclaration(d.typ, "Mark", provider.Mark)
	} else {
		provider := surrogate.(decl.FamilyProvider)
		family, err = capability.InvokeDeclaration(d.typ, "Family", provider.Family)
	}
	if err != nil {
		return err
	}
	if family == "" && d.role != legendRole {
		return fmt.Errorf("%v: only a legend may declare an empty mark family", d.typ)
	}

	var discriminants []string
	if provider, ok := surrogate.(decl.DiscriminantProvider); ok {
		discriminants, err = capability.InvokeDeclaration(d.typ, "Discriminate", provider.Discriminate)
		if err != nil {
			return err
		}
	}
	d.mark, err = newMark(tag, family, discriminants)
	if err != nil {
		return fmt.Errorf("%v: %w", d.typ, err)
	}
	return nil
}

func (d *markerDescriptor) collectFeature(field reflect.StructField, seen map[string]bool) error {
	if !reflectutils.FieldImplements[decl.FeatureDeclaration](field) {
		return fmt.Errorf("%v.%s: feature must embed legend.Feature", d.typ, field.Name)
	}
	symbol, err := tags.ReadSingle(field.Tag, "symbol", validateSymbol)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", d.typ, field.Name, err)
	}
	if seen[symbol] {
		return fmt.Errorf("%v.%s: duplicate symbol %q", d.typ, field.Name, symbol)
	}
	path, err := reflectutils.EmbeddedPath(field.Type, reflect.TypeFor[decl.Feature]())
	if err != nil {
		return fmt.Errorf("%v.%s: %w", d.typ, field.Name, err)
	}
	seen[symbol] = true
	d.features = append(d.features, featureDescriptor{symbol: symbol, declarationPath: slices.Concat(field.Index, path)})
	return nil
}

// candidates preserves occurrence identity while sharing target descriptors.
// The traversal policy is responsible for choosing an eligible lane.
func (d *markerDescriptor) candidates(laneIndex int) []candidate {
	spots := d.lanes[laneIndex].spots
	result := make([]candidate, len(spots))
	for i, spot := range spots {
		result[i] = candidate{laneIndex: laneIndex, spotIndex: i, target: spot.target}
	}
	return result
}
