package sail

import (
	"fmt"
	"reflect"

	"github.com/Qlustra/sail/internals/utilities"
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
	typ        reflect.Type
	role       locationRole
	mark       mark
	qualifiers []qualifier
	features   []feature
	lanes      []lane
	behavior   Behavior
}

type feature struct {
	symbol      string
	schemaField []int
}

type qualifier struct {
	symbol      string
	schemaField []int
}

// newLocation starts a local description; the explorer records it only after
// its fields, mark, and behavior have passed validation.
func newLocation(t reflect.Type, role locationRole) *location {
	return &location{typ: t, role: role}
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
	seen[symbol] = true
	l.features = append(l.features, feature{symbol: symbol, schemaField: append([]int(nil), field.Index...)})
	return nil
}

func (l *location) collectQualifier(field reflect.StructField, seen map[string]bool) error {
	if !utilities.FieldImplements[qualifierDeclaration](field) {
		return fmt.Errorf("%v.%s: qualifier must embed sail.Qualifier", l.typ, field.Name)
	}
	symbol, err := readSingleTag(field.Tag, "symbol", validateSymbol)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", l.typ, field.Name, err)
	}
	if seen[symbol] {
		return fmt.Errorf("%v.%s: duplicate symbol %q", l.typ, field.Name, symbol)
	}
	seen[symbol] = true
	l.qualifiers = append(l.qualifiers, qualifier{symbol: symbol, schemaField: append([]int(nil), field.Index...)})
	return nil
}

func (l *location) collectBehavior(surrogate any) error {
	if provider, ok := surrogate.(BehaviorProvider); ok {
		behavior, err := utilities.InvokeDeclaration(l.typ, "Behavior", provider.Behavior)
		if err != nil {
			return err
		}
		l.behavior = append(Behavior(nil), behavior...)
		for _, capability := range l.behavior {
			if capability.typ == nil || capability.typ.Kind() != reflect.Interface {
				return fmt.Errorf("%v: behavior capability must describe an interface", l.typ)
			}
		}
	}
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
