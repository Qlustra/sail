package sail

import "errors"

// Course will describe a live application traversal.
type Course struct{}

// ErrTraversalNotImplemented indicates that live traversal and hydration are
// not available yet. Chart compilation and field surveying are implemented.
var ErrTraversalNotImplemented = errors.New("sail: live application traversal is not implemented")

func (c *Chart[L]) Traverse(app any, course *Course) error {
	return ErrTraversalNotImplemented
}
