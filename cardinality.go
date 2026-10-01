package sail

import (
	"fmt"
	"reflect"
)

type cardinality struct {
	min int
	max int // -1 means unbounded.
}

func compileCardinality(kind spotKind, tags reflect.StructTag) (cardinality, error) {
	if kind != multiSpot {
		if _, present := tags.Lookup("multiplicity"); present {
			return cardinality{}, fmt.Errorf("multiplicity is only valid on MultiSpot")
		}
	}

	switch kind {
	case singleSpot:
		return cardinality{min: 1, max: 1}, nil
	case optionalSpot:
		return cardinality{min: 0, max: 1}, nil
	case multiSpot:
		value, err := readSingleTag(tags, "multiplicity", nil)
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
