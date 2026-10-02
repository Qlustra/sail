package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

type locationRole string

const (
	chartRole    locationRole = "chart"
	landmarkRole locationRole = "landmark"
	regionRole   locationRole = "region"
)

// A location describes a reusable schema type. Lanes and spots describe
// occurrences within it; none of these descriptors hold live subject values.
type location struct {
	typ          reflect.Type
	role         locationRole
	mark         mark
	acceptance   Acceptance
	qualifiers   []qualifier
	features     []feature
	barrels      []barrel
	lanes        []lane
	carrierField []int
}

type feature struct {
	symbol      string
	schemaField []int
}

// newLocation starts a local description; the explorer records it only after
// its fields, mark, and acceptance have passed validation.
func newLocation(t reflect.Type, role locationRole) (*location, error) {
	var carrier reflect.Type
	switch role {
	case chartRole:
		carrier = reflect.TypeFor[Chart]()
	case regionRole:
		carrier = reflect.TypeFor[Region]()
	case landmarkRole:
		carrier = reflect.TypeFor[Landmark]()
	}
	path, err := carrierPath(t, carrier)
	if err != nil {
		return nil, err
	}
	return &location{typ: t, role: role, carrierField: path}, nil
}

func (l *location) collectMark(tag string, surrogate any) error {
	var family string
	var err error
	// The explorer checks required providers before allocating the surrogate.
	if l.role == chartRole {
		provider := surrogate.(MarkProvider)
		tag, err = utilities.InvokeDeclaration(l.typ, "Mark", provider.Mark)
	} else {
		provider := surrogate.(FamilyProvider)
		family, err = utilities.InvokeDeclaration(l.typ, "Family", provider.Family)
	}
	if err != nil {
		return err
	}
	if family == "" && l.role != chartRole {
		return fmt.Errorf("%v: only a chart may declare an empty mark family", l.typ)
	}

	var discriminants []string
	if provider, ok := surrogate.(DiscriminantProvider); ok {
		discriminants, err = utilities.InvokeDeclaration(l.typ, "Discriminate", provider.Discriminate)
		if err != nil {
			return err
		}
	}
	l.mark, err = newMark(tag, family, discriminants)
	if err != nil {
		return fmt.Errorf("%v: %w", l.typ, err)
	}
	return nil
}

func (l *location) collectFeature(field reflect.StructField, seen map[string]bool) error {
	if !utilities.FieldImplements[featureDeclaration](field) {
		return fmt.Errorf("%v.%s: feature must embed sail.Feature", l.typ, field.Name)
	}
	symbol, err := readSingleTag(field.Tag, "symbol", validateSymbol)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", l.typ, field.Name, err)
	}
	if seen[symbol] {
		return fmt.Errorf("%v.%s: duplicate symbol %q", l.typ, field.Name, symbol)
	}
	path, err := carrierPath(field.Type, reflect.TypeFor[Feature]())
	if err != nil {
		return fmt.Errorf("%v.%s: %w", l.typ, field.Name, err)
	}
	seen[symbol] = true
	l.features = append(l.features, feature{symbol: symbol, schemaField: slices.Concat(field.Index, path)})
	return nil
}

// candidates preserves occurrence identity while sharing target locations.
// The traversal policy is responsible for choosing an eligible lane.
func (l *location) candidates(laneIndex int) []candidate {
	spots := l.lanes[laneIndex].spots
	result := make([]candidate, len(spots))
	for i, spot := range spots {
		result[i] = candidate{laneIndex: laneIndex, spotIndex: i, target: spot.target}
	}
	return result
}
