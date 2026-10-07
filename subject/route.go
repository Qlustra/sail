package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Route determines movement after an interceptor handles an occurrence.
// Construct one with Pass, Descend, DescendSelected, Prune, or Stop.
type Route = subjectcore.Route

// Pass declines this interceptor and offers the occurrence to the next one.
func Pass() Route { return subjectcore.Pass() }

// Descend accepts the occurrence and visits its children in the chosen order.
// An optional completion callback runs after its descendant scope succeeds.
// At most one completion callback may be supplied; nil means no callback.
// Invalid arguments are reported by propagation rather than panicking here.
func Descend(order Order, complete ...func() error) Route {
	return subjectcore.Descend(order, complete...)
}

// DescendSelected accepts the occurrence and visits only the supplied typed
// children in the chosen order. The source replaces both declared children and
// Enter() for this occurrence; it does not change the carrier or its cache.
// Selected members use the Vessel's Legend and join the same propagation.
// Order and completion follow Descend's rules. Invalid arguments are reported
// by propagation; a valid empty source completes without visiting children.
func DescendSelected(order Order, children Children, complete ...func() error) Route {
	return subjectcore.DescendSelected(order, children, complete...)
}

// Prune accepts the occurrence, skips its children, and continues siblings.
func Prune() Route { return subjectcore.Prune() }

// Stop finishes the entire propagation successfully.
func Stop() Route { return subjectcore.Stop() }

// Abort returns a stopping route and the original error for direct forwarding
// from an interceptor. A nil error has the same effect as Stop.
func Abort(err error) (Route, error) { return subjectcore.Abort(err) }
