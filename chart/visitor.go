package chart

import chartcore "github.com/qlustra/sail/internal/chart"

// Visitor processes one selected occurrence. Returning an error ends execution.
// Use On to receive a typed marker alongside its encounter context.
type Visitor = chartcore.Visitor

// Visitors processes each encounter in registration order. All matching typed
// visitors participate; the first error ends both the list and the traversal.
type Visitors = chartcore.Visitors

// Visit creates a single-item visitor list from a typed callback.
func Visit[T any](visit func(T, Encounter) error) Visitors {
	return chartcore.Visit(visit)
}

// On adapts a typed callback to one visitor for use in a Visitors list.
// Unmatched markers do not call visit; they remain part of the route.
// Pointer callbacks receive original Chart values. Interface callbacks prefer
// the marker's pointer over its value.
func On[T any](visit func(T, Encounter) error) Visitor {
	return chartcore.On(visit)
}
