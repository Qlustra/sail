package sail

import (
	"fmt"

	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/traversal"
)

// TraversalOrder orders siblings in a survey or Course search, not rules or
// fields in the populated Legend.
type TraversalOrder string

const (
	// ChartOrder encounters siblings in Chart declaration order, including
	// promoted properties at their embedding positions.
	ChartOrder TraversalOrder = "Chart order"
	// LegendOrder encounters siblings in Legend Lane and Spot declaration
	// order; repeated occurrences in one Spot retain their Chart order.
	LegendOrder TraversalOrder = "Legend order"
)

// Encounter describes the incoming grammatical step, not the selected token.
type Encounter struct {
	depth int
	lane  *declarations.LaneCarrier
}

// Depth counts token steps from the Legend root. Its immediate tokens have depth 1.
func (e Encounter) Depth() int { return e.depth }

// LaneID is empty for an unnamed Lane. IDs are local to the enclosing body.
func (e Encounter) LaneID() string {
	if e.lane == nil {
		return ""
	}
	return e.lane.ID
}

// SpotSelector filters tokens by their stored pointer type, optionally by Lane
// and predicate. Its zero value selects by type only.
type SpotSelector[T any] struct {
	predicate func(T) bool
	laneID    string
	err       error
}

// Select narrows the actual stored token pointer to T, optionally by one
// non-nil predicate. T may be a pointer or an interface implemented by that
// pointer; Sail does not copy the token to match a non-pointer struct type.
// An invalid predicate list is reported by the operation using the selector.
func Select[T any](predicates ...func(T) bool) SpotSelector[T] {
	switch len(predicates) {
	case 0:
		return SpotSelector[T]{}
	case 1:
		if predicates[0] != nil {
			return SpotSelector[T]{predicate: predicates[0]}
		}
	}
	return SpotSelector[T]{err: fmt.Errorf("Select requires at most one non-nil predicate")}
}

// InLane restricts eligibility before the predicate runs, without pruning descent.
func (s SpotSelector[T]) InLane(id string) SpotSelector[T] {
	if id == "" {
		s.err = fmt.Errorf("InLane requires a nonempty Lane ID")
	}
	s.laneID = id
	return s
}

// Rule is a typed Survey encounter registered with SpotSelector.Encounter.
type Rule[P any] struct {
	apply func(Encounter, P, any) P
	err   error
}

// isTheSpotWereLookingFor tests a candidate without running its encounter. A survey can reduce a
// match immediately; a waypoint can check for ambiguity before accepting it.
func (s SpotSelector[T]) isTheSpotWereLookingFor(at Encounter, candidateMarkerPointer any) (T, bool) {
	typedMarker, markerIsTheTypeWeWant := candidateMarkerPointer.(T)
	spotIsInTheLaneWeWant := s.laneID == "" || at.LaneID() == s.laneID
	if !markerIsTheTypeWeWant || !spotIsInTheLaneWeWant || (s.predicate != nil && !s.predicate(typedMarker)) {
		var zero T
		return zero, false
	}
	return typedMarker, true
}

// Encounter uses reduce on matching tokens. Its result is the prefix passed
// to the next rule at that token, or to its children if it is the last rule.
func (s SpotSelector[T]) Encounter[P any](reduce func(Encounter, P, T) P) Rule[P] {
	if s.err != nil {
		return Rule[P]{err: s.err}
	}
	if reduce == nil {
		return Rule[P]{err: fmt.Errorf("missing encounter callback")}
	}
	return Rule[P]{apply: func(at Encounter, inherited P, candidateMarkerPointer any) P {
		token, ok := s.isTheSpotWereLookingFor(at, candidateMarkerPointer)
		if !ok {
			return inherited
		}
		return reduce(at, inherited, token)
	}}
}

// Survey visits present tokens depth-first below the populated Legend (the
// depth-zero root), in the given sibling order. Matching reducers compose in
// registration order. Each child receives its parent's final prefix by value:
// sibling replacements do not leak, referenced storage is shared. Matching
// filters encounters, not descent. Survey does not collect a result.
func (c *Navigator[L]) Survey[P any](legend *L, order TraversalOrder, scopeSeed P, rules ...Rule[P]) error {
	if c == nil || c.grammar == nil || legend == nil {
		return fmt.Errorf("Survey requires a compiled Navigator and a populated Legend")
	}
	if order != ChartOrder && order != LegendOrder {
		return fmt.Errorf("invalid Survey order %q", order)
	}
	for i, rule := range rules {
		if rule.err != nil {
			return fmt.Errorf("rule %d: %w", i+1, rule.err)
		}
		if rule.apply == nil {
			return fmt.Errorf("rule %d: missing encounter callback", i+1)
		}
	}
	return traversal.Survey(c.grammar, legend, order == ChartOrder, scopeSeed, func(depth int, lane *declarations.LaneCarrier, candidateMarkerPointer any, inheritedScope P) P {
		at := Encounter{depth: depth, lane: lane}
		for _, rule := range rules {
			inheritedScope = rule.apply(at, inheritedScope, candidateMarkerPointer)
		}
		return inheritedScope
	})
}
