package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

type lane struct {
	schemaField []int
	strategy    LaneExplorationStrategy
	spots       []spot // Alternatives within the lane; their order is not policy.
}

// nullable reports whether the lane can be skipped without any occurrences.
func (l lane) nullable() bool {
	for _, spot := range l.spots {
		if spot.cardinality.min > 0 {
			return false
		}
	}
	return true
}

// laneField retains the local inspection needed for later lane expansion.
type laneField struct {
	name        string
	typ         reflect.Type
	schemaField []int
	strategy    LaneExplorationStrategy
}

func inspectLane(field reflect.StructField) (laneField, error) {
	if !isLaneField(field) {
		return laneField{}, fmt.Errorf("lane must embed sail.Lane")
	}
	typ := utilities.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct {
		return laneField{}, fmt.Errorf("lane must resolve to a struct")
	}
	strategy, err := readLaneStrategy(field.Tag)
	if err != nil {
		return laneField{}, err
	}
	return laneField{
		name:        field.Name,
		typ:         typ,
		schemaField: append([]int(nil), field.Index...),
		strategy:    strategy,
	}, nil
}

func readLaneStrategy(tags reflect.StructTag) (LaneExplorationStrategy, error) {
	format := SingleValueTag{Tag: Tag{Name: "exploration"}}
	present, err := readTag(tags, "exploration", &format)
	if err != nil {
		return "", err
	}
	if !present {
		return ExploreUnambiguously, nil
	}
	strategy := LaneExplorationStrategy(format.Raw)
	switch strategy {
	case ExploreEagerly, ExploreUnambiguously:
		return strategy, nil
	default:
		return "", fmt.Errorf("invalid exploration strategy %q", format.Raw)
	}
}
