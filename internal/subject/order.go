package subject

import "fmt"

// Order determines how propagation visits an occurrence's descendants.
// An order must be chosen explicitly; its zero value is invalid.
type Order uint8

const (
	// DepthFirst visits each branch through completion before the next sibling.
	DepthFirst Order = iota + 1
	// BreadthFirst visits siblings before their descendants.
	BreadthFirst
)

func (order Order) validate() error {
	switch order {
	case DepthFirst, BreadthFirst:
		return nil
	default:
		return fmt.Errorf("invalid propagation order %d; choose DepthFirst or BreadthFirst", order)
	}
}
