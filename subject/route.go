package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Route determines movement after an interceptor handles an occurrence.
// Construct one with Pass, Descend, Prune, or Stop.
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

// Prune accepts the occurrence, skips its children, and continues siblings.
func Prune() Route { return subjectcore.Prune() }

// Stop finishes the entire propagation successfully.
func Stop() Route { return subjectcore.Stop() }

// Abort returns a stopping route and the original error for direct forwarding
// from an interceptor. A nil error has the same effect as Stop.
func Abort(err error) (Route, error) { return subjectcore.Abort(err) }
