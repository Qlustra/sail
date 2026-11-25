package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/traversal"
)

// A Plot is an editable sequence of Waypoints for one Navigator. It cannot run.
type Plot[L any] struct {
	navigator *Navigator[L]
	steps     []waypoint
}

// Course holds a checked sequence, independent of subsequent edits to its Plot.
type Course[L any] struct {
	navigator *Navigator[L]
	steps     []traversal.WaypointMatch
}

type waypoint struct {
	shape traversal.WaypointShape
	err   error
	match traversal.WaypointMatch
}

// PlotCourse starts an empty Plot for this Navigator. Add waypoints and call
// Prepare before searching a populated Legend.
func (c *Navigator[L]) PlotCourse() *Plot[L] { return &Plot[L]{navigator: c} }

// Waypoint adds a selector by value. Each step retains its own marker type.
func (p *Plot[L]) Waypoint[T any](s SpotSelector[T]) {
	p.steps = append(p.steps, waypoint{
		shape: traversal.WaypointShape{Marker: reflect.TypeFor[T](), LaneID: s.laneID},
		err:   s.err,
		match: func(depth int, lane *declarations.LaneCarrier, marker any) (any, bool) {
			return s.isTheSpotWereLookingFor(Encounter{depth: depth, lane: lane}, marker)
		},
	})
}

// Prepare checks the selector configuration and the existence of a connected
// path in the grammar. It never invokes predicates or inspects a populated Legend.
func (p *Plot[L]) Prepare() (*Course[L], error) {
	if p == nil || p.navigator == nil || p.navigator.grammar == nil {
		return nil, fmt.Errorf("prepare Course requires a compiled Navigator")
	}
	if len(p.steps) == 0 {
		return nil, fmt.Errorf("prepare Course requires at least one waypoint")
	}
	shapes := make([]traversal.WaypointShape, len(p.steps))
	steps := make([]traversal.WaypointMatch, len(p.steps))
	for i, step := range p.steps {
		if step.err != nil {
			return nil, fmt.Errorf("waypoint %d: %w", i+1, step.err)
		}
		shapes[i] = step.shape
		steps[i] = step.match
	}
	if err := traversal.ValidateWaypoints(p.navigator.grammar, shapes); err != nil {
		return nil, err
	}
	return &Course[L]{navigator: p.navigator, steps: steps}, nil
}
