package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/engine"
)

// Traversal selects a connected route through a populated World and processes
// its encounters. Course must contain at least one waypoint; Visit is required.
type Traversal struct {
	Course Course
	Visit  Visitor
}

// Visitor processes one selected occurrence. Returning an error ends execution.
// Use On to receive a typed marker alongside its encounter context.
type Visitor func(Encounter) error

// On adapts a typed callback to a visitor. Unmatched markers do not call visit;
// they remain part of the route. Pointer callbacks receive original World
// storage. Interface callbacks prefer the marker's pointer over its value.
func On[T any](visit func(T, Encounter) error) Visitor {
	if visit == nil {
		return nil
	}
	return func(encounter Encounter) error {
		if encounter.occurrence == nil {
			return fmt.Errorf("sail: visitor requires a World encounter")
		}
		marker, matches := markerAs[T](encounter.occurrence.Value)
		if !matches {
			return nil
		}
		return visit(marker, encounter)
	}
}

// Traverse resolves the entire Course against world before invoking any visitor.
// Every waypoint must select exactly one immediate child marker. The Legend root
// anchors navigation but is not itself visited. Selected occurrences are visited
// in route order, ending at the destination.
//
// Selectors must be read-only and World topology must remain stable throughout
// planning and execution. Sail does not detect mutations. Callback panics
// propagate; returned visitor errors retain their cause and encounter context.
// This operation neither traverses a live Application nor invokes its hooks.
func (c *Chart[L]) Traverse(world *L, traversal Traversal) error {
	if traversal.Visit == nil {
		return fmt.Errorf("sail: traverse requires a visitor")
	}
	course := make([]engine.Waypoint, len(traversal.Course))
	for i, waypoint := range traversal.Course {
		if waypoint.err != nil {
			return fmt.Errorf("sail: waypoint %d: %w", i+1, waypoint.err)
		}
		course[i] = waypoint.selection
	}
	var compiled *engine.Chart
	if c != nil {
		compiled = c.compiled
	}
	plan, err := engine.PlanRoute(compiled, reflect.ValueOf(world), course)
	if err != nil {
		return err
	}
	for i, occurrence := range plan.Occurrences {
		encounter := Encounter{occurrence: occurrence, step: i, length: len(plan.Occurrences)}
		if err := traversal.Visit(encounter); err != nil {
			return fmt.Errorf("sail: waypoint %d at %s: %w", i+1, occurrence.Source, err)
		}
	}
	return nil
}
