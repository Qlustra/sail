package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/tags"
)

type cardinality struct {
	min int
	max int // -1 means unbounded.
}

func compileCardinality(kind decl.SpotKind, structTags reflect.StructTag) (cardinality, error) {
	if kind != decl.MultiSpotKind {
		if _, present := structTags.Lookup("multiplicity"); present {
			return cardinality{}, fmt.Errorf("multiplicity is only valid on MultiSpot")
		}
	}

	switch kind {
	case decl.SingleSpotKind:
		return cardinality{min: 1, max: 1}, nil
	case decl.OptionalSpotKind:
		return cardinality{min: 0, max: 1}, nil
	case decl.MultiSpotKind:
		value, err := tags.ReadSingle(structTags, "multiplicity", nil)
		if err != nil {
			return cardinality{}, err
		}
		switch value {
		case "any":
			return cardinality{min: 0, max: -1}, nil
		case "some":
			return cardinality{min: 1, max: -1}, nil
		default:
			return cardinality{}, fmt.Errorf("invalid multiplicity %q: expected any or some", value)
		}
	default:
		return cardinality{}, fmt.Errorf("invalid spot kind %q", kind)
	}
}
