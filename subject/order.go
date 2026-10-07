package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Order determines how propagation visits an occurrence's descendants.
// An order must be chosen explicitly; its zero value is invalid.
type Order = subjectcore.Order

const (
	// DepthFirst visits each branch through completion before the next sibling.
	DepthFirst = subjectcore.DepthFirst
	// BreadthFirst visits siblings before their descendants.
	BreadthFirst = subjectcore.BreadthFirst
)
