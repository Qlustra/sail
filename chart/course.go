package chart

import chartcore "github.com/qlustra/sail/internal/chart"

// Course is a reusable sequence of immediate-child selections starting at the
// Chart root. It retains neither matched markers nor execution progress.
type Course = chartcore.Course

// Waypoint selects the next marker in a connected Course. Construct one with
// Select. A zero Waypoint is invalid and is rejected before visiting begins.
type Waypoint = chartcore.Waypoint

// Select matches a marker value, pointer, or interface T, optionally narrowed
// by one predicate. Interface matching prefers the marker's pointer.
//
// A predicate receives a shallow copy of the concrete marker (a pointer to that
// copy when T is a pointer or a matching interface). Referenced data remains
// shared under ordinary Go copy semantics. Predicates must be read-only; Sail
// neither deep-copies the Chart nor detects mutations. Invalid predicate lists
// are reported by Traverse rather than panicking here.
func Select[T any](predicates ...func(T) bool) Waypoint {
	return chartcore.Select(predicates...)
}
