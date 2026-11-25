package sail

import (
	"fmt"

	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/traversal"
)

// SpotSelection retains the matched token and its grammatical position, so another
// selector can start there without searching the populated Legend again.
type SpotSelection[T any] struct {
	marker   T
	position traversal.SpotPosition
}

// Marker returns the stored token pointer (as T), or the zero value of T if
// the step did not uniquely match.
func (s SpotSelection[T]) Marker() T { return s.marker }

// Matched reports that exactly one immediate token matched.
func (s SpotSelection[T]) Matched() bool { return s.position.Matched() }

// Ambiguous reports that more than one immediate token matched.
func (s SpotSelection[T]) Ambiguous() bool { return s.position.Ambiguous() }

// Next tests the immediate markers of the populated Legend. After checking err,
// Matched means exactly one marker was returned; Ambiguous means more than one
// matched. If both are false, there was no match. Only a unique selection can advance.
func (c *Navigator[L]) Next[T any](legend *L, selector SpotSelector[T]) (SpotSelection[T], T, error) {
	if c == nil {
		sel := SpotSelection[T]{}
		return sel, sel.marker, fmt.Errorf("selection requires a compiled Navigator")
	}
	start, err := traversal.Start(c.grammar, legend)
	if err != nil {
		sel := SpotSelection[T]{}
		return sel, sel.marker, err
	}
	return selector.next(start)
}

// Next takes the same step from a selected token. Its type can differ from the
// preceding step's type. It never searches through a nonmatching Region.
func (s SpotSelection[T]) Next[U any](selector SpotSelector[U]) (SpotSelection[U], U, error) {
	return selector.next(s.position)
}

func (s SpotSelector[T]) next(from traversal.SpotPosition) (SpotSelection[T], T, error) {
	if s.err != nil {
		sel := SpotSelection[T]{}
		return sel, sel.marker, s.err
	}
	position, marker, err := from.Next(func(depth int, lane *declarations.LaneCarrier, candidateMarkerPointer any) (T, bool) {
		return s.isTheSpotWereLookingFor(Encounter{depth: depth, lane: lane}, candidateMarkerPointer)
	})
	if err != nil {
		sel := SpotSelection[T]{}
		return sel, sel.marker, err
	}
	sel := SpotSelection[T]{marker: marker, position: position}
	return sel, sel.marker, nil
}
