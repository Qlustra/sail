package subject

import "fmt"

type routeAction uint8

const (
	passRoute routeAction = iota
	descendRoute
	pruneRoute
	stopRoute
)

// Route determines movement after an interceptor handles an occurrence.
// Construct one with Pass, Descend, DescendSelected, Prune, or Stop.
type Route struct {
	action   routeAction
	order    Order
	complete func() error
	selected *Children
	err      error
}

// Pass declines this interceptor and offers the occurrence to the next one.
func Pass() Route { return Route{} }

// Descend accepts the occurrence and visits its children in the chosen order.
// An optional completion callback runs after its descendant scope succeeds.
// At most one completion callback may be supplied; nil means no callback.
// Invalid arguments are reported by propagation rather than panicking here.
func Descend(order Order, complete ...func() error) Route {
	return descent("Descend", order, complete)
}

// DescendSelected replaces this occurrence's normal children with the supplied
// typed source. Ordering and successful completion follow Descend's rules.
func DescendSelected(order Order, children Children, complete ...func() error) Route {
	route := descent("DescendSelected", order, complete)
	if route.err == nil {
		route.err = children.validate()
	}
	route.selected = &children
	return route
}

func descent(name string, order Order, complete []func() error) Route {
	route := Route{action: descendRoute, order: order}
	if err := order.validate(); err != nil {
		route.err = err
		return route
	}
	if len(complete) > 1 {
		route.err = fmt.Errorf("%s accepts at most one completion callback, got %d", name, len(complete))
		return route
	}
	if len(complete) == 1 {
		route.complete = complete[0]
	}
	return route
}

// Prune accepts the occurrence, skips its children, and continues siblings.
func Prune() Route { return Route{action: pruneRoute} }

// Stop finishes the entire propagation successfully.
func Stop() Route { return Route{action: stopRoute} }

// Abort returns a stopping route and the original error for direct forwarding
// from an interceptor. A nil error has the same effect as Stop.
func Abort(err error) (Route, error) { return Stop(), err }

func (route Route) validate() error {
	if route.err != nil {
		return route.err
	}
	switch route.action {
	case passRoute, pruneRoute, stopRoute:
		return nil
	case descendRoute:
		return route.order.validate()
	default:
		return fmt.Errorf("invalid route action %d", route.action)
	}
}
