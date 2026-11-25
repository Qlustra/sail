package traversal

import (
	"errors"
	"fmt"
	"iter"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
)

// SpotPosition pairs a stored Legend value with the Lane uses for its immediate
// tokens. A selected Landmark has a value but no Lanes.
type SpotPosition struct {
	markerValue *capture.ValueInterrogator
	lanes       []*declarations.LaneCarrier
	depth       int
	matched     bool
	ambiguous   bool
}

func (s SpotPosition) Matched() bool { return s.matched }

func (s SpotPosition) Ambiguous() bool { return s.ambiguous }

func Start[L any](grammar *declarations.Legend, legend *L) (SpotPosition, error) {
	if grammar == nil || legend == nil {
		return SpotPosition{}, errors.New("selection requires a compiled Chart and a populated Legend")
	}
	value, check := capture.Value(legend)
	if !check.OK() {
		return SpotPosition{}, fmt.Errorf("select from Legend: %s", check.Why())
	}
	return SpotPosition{markerValue: value, lanes: grammar.Lanes}, nil
}

func (p SpotPosition) Next[T any](match func(int, *declarations.LaneCarrier, any) (T, bool)) (SpotPosition, T, error) {
	var zero T
	matches, err := p.immediateMatches(match, false)
	if err != nil {
		return SpotPosition{}, zero, err
	}
	var next SpotPosition
	var selected T
	for position, marker := range matches {
		if next.Matched() {
			return SpotPosition{ambiguous: true}, zero, nil
		}
		next, selected = position, marker
	}
	return next, selected, nil
}

// immediateMatches captures this body's tokens in the requested order. Matching is
// done on iteration: the consumer can descend from a yielded position before
// the next sibling is tested, or break without evaluating further predicates.
func (p SpotPosition) immediateMatches[T any](match func(int, *declarations.LaneCarrier, any) (T, bool), chartOrder bool) (iter.Seq2[SpotPosition, T], error) {
	if p.markerValue == nil {
		return nil, errors.New("select from an uninitialized position")
	}
	marker, check := p.markerValue.AsStruct()
	if !check.OK() {
		return nil, fmt.Errorf("select from marker: %s", check.Why())
	}
	spotTokens, err := getBodySpotTokens(marker, p.lanes, chartOrder)
	if err != nil {
		return nil, err
	}
	return func(yield func(SpotPosition, T) bool) {
		for _, candidate := range spotTokens {
			marker, ok := match(p.depth+1, candidate.lane, candidate.markerPointer)
			if !ok {
				continue
			}
			position := SpotPosition{markerValue: candidate.markerValue, depth: p.depth + 1, matched: true}
			if candidate.spot.For == declarations.SpotForRegion {
				position.lanes = candidate.spot.Region.Declaration.Lanes
			}
			if !yield(position, marker) {
				return
			}
		}
	}, nil
}
