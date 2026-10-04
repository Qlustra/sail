package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
	"github.com/qlustra/sail/internal/tags"
)

type laneDescriptor struct {
	declarationPath []int
	strategy        decl.LaneExplorationStrategy
	spots           []spotDescriptor // Alternatives within the lane; their order is not policy.
}

// nullable reports whether the lane can be skipped without any occurrences.
func (l laneDescriptor) nullable() bool {
	for _, spot := range l.spots {
		if spot.cardinality.min > 0 {
			return false
		}
	}
	return true
}

// laneField retains the local inspection needed for later lane expansion.
type laneField struct {
	name            string
	typ             reflect.Type
	declarationPath []int
	strategy        decl.LaneExplorationStrategy
}

func inspectLane(field reflect.StructField) (laneField, error) {
	if !reflectutils.FieldImplements[decl.LaneDeclaration](field) {
		return laneField{}, fmt.Errorf("lane must embed legend.Lane")
	}
	typ := reflectutils.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct {
		return laneField{}, fmt.Errorf("lane must resolve to a struct")
	}
	strategy, err := readLaneStrategy(field.Tag)
	if err != nil {
		return laneField{}, err
	}
	return laneField{
		name:            field.Name,
		typ:             typ,
		declarationPath: append([]int(nil), field.Index...),
		strategy:        strategy,
	}, nil
}

func readLaneStrategy(structTags reflect.StructTag) (decl.LaneExplorationStrategy, error) {
	format := tags.SingleValue{Tag: tags.Tag{Name: "exploration"}}
	present, err := tags.Read(structTags, "exploration", &format)
	if err != nil {
		return "", err
	}
	if !present {
		return decl.ExploreUnambiguously, nil
	}
	strategy := decl.LaneExplorationStrategy(format.Raw)
	switch strategy {
	case decl.ExploreEagerly, decl.ExploreUnambiguously:
		return strategy, nil
	default:
		return "", fmt.Errorf("invalid exploration strategy %q", format.Raw)
	}
}
