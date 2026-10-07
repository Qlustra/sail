package engine

import (
	"fmt"
	"reflect"
	"strings"
)

// Waypoint selects exactly one immediate child. Match owns type matching;
// Type names the requested callback type in diagnostics.
type Waypoint struct {
	Type  reflect.Type
	Match func(reflect.Value) bool
}

// RoutePlan contains only selected markers, in Course order. The first marker's
// Parent retains the Chart root so processors can inspect the complete route.
type RoutePlan struct {
	Occurrences []*Occurrence
}

// PlanRoute resolves the whole Course before any processor can run. Selectors
// inspect existing Chart values; traversal itself performs no allocations into
// the Chart and invokes no Subject hooks.
func PlanRoute(charter *Charter, chart reflect.Value, course []Waypoint) (*RoutePlan, error) {
	parent, err := navigationRoot(charter, chart)
	if err != nil {
		return nil, err
	}
	if len(course) == 0 {
		return nil, fmt.Errorf("sail: traversal requires a nonempty Course")
	}
	for index, waypoint := range course {
		if waypoint.Type == nil || waypoint.Match == nil {
			return nil, fmt.Errorf("sail: waypoint %d (%v) requires a type and selector", index+1, waypoint.Type)
		}
	}
	plan := &RoutePlan{Occurrences: make([]*Occurrence, 0, len(course))}
	for index, waypoint := range course {
		selected, err := selectChild(parent, waypoint)
		if err != nil {
			return nil, fmt.Errorf("sail: waypoint %d (%v) under %s: %w", index+1, waypoint.Type, parent.Source, err)
		}
		plan.Occurrences = append(plan.Occurrences, selected)
		parent = selected
	}
	return plan, nil
}

func selectChild(parent *Occurrence, waypoint Waypoint) (*Occurrence, error) {
	children, err := immediateChildren(parent)
	if err != nil {
		return nil, err
	}
	var matches []*Occurrence
	for _, child := range children {
		if waypoint.Match(child.Value) {
			matches = append(matches, child)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no immediate child matches")
	}
	if len(matches) > 1 {
		paths := make([]string, len(matches))
		for index, match := range matches {
			paths[index] = match.Source.String()
		}
		return nil, fmt.Errorf("ambiguous selector matches %d immediate children: %s", len(matches), strings.Join(paths, ", "))
	}
	return matches[0], nil
}
