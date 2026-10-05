package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/engine"
)

// Traversal selects a connected route through a populated World and processes
// its encounters. Course and Visit must each contain at least one entry.
type Traversal struct {
	Course Course
	Visit  Visitors
}

// Traverse resolves the entire Course against world before invoking any visitor.
// Every waypoint must select exactly one immediate child marker. The Legend root
// anchors navigation but is not itself visited. Selected occurrences are visited
// in route order, ending at the destination. Each encounter passes through every
// registered visitor in list order. Visitor registrations are copied on entry.
//
// Selectors must be read-only and World topology must remain stable throughout
// planning and execution. Sail does not detect mutations. Callback panics
// propagate; returned visitor errors retain their cause and encounter context.
// This operation neither traverses a live Application nor invokes its hooks.
func (c *Chart[L]) Traverse(world *L, traversal Traversal) error {
	visitors := slices.Clone(traversal.Visit)
	if err := visitors.validate(); err != nil {
		return fmt.Errorf("sail: traverse: %w", err)
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
		if err := visitors.visit(encounter); err != nil {
			return fmt.Errorf("sail: waypoint %d at %s: %w", i+1, occurrence.Source, err)
		}
	}
	return nil
}
