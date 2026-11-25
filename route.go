package sail

import (
	"errors"
	"fmt"

	"github.com/qlustra/sail/internal/chart/traversal"
)

// Route contains one stored marker pointer per waypoint in Course order. Its
// slice is independent of other routes; the markers belong to the populated Legend.
type Route []any

// Cardinality errors describe complete routes, never an ambiguous intermediate step.
var (
	ErrNoRoute        = errors.New("no complete route matches")
	ErrAmbiguousRoute = errors.New("more than one complete route matches")
)

// FirstRoute returns the first complete route in depth-first order using the
// requested sibling order. A matching prefix alone does not end the search.
func (c *Course[L]) FirstRoute(legend *L, order TraversalOrder) (Route, error) {
	return c.singleRoute(legend, order, false)
}

// RequireOneRoute returns the sole complete route, ErrNoRoute, or ErrAmbiguousRoute.
func (c *Course[L]) RequireOneRoute(legend *L, order TraversalOrder) (Route, error) {
	return c.singleRoute(legend, order, true)
}

func (c *Course[L]) singleRoute(legend *L, order TraversalOrder, requireUnique bool) (Route, error) {
	var selected Route
	var ambiguous bool
	err := c.routes(legend, order, func(route []any) bool {
		if selected != nil {
			ambiguous = true
			return false
		}
		selected = route
		return requireUnique // Only a uniqueness check needs to look beyond the first route.
	})
	switch {
	case err != nil:
		return nil, err
	case ambiguous:
		return nil, ErrAmbiguousRoute
	case selected == nil:
		return nil, ErrNoRoute
	default:
		return selected, nil
	}
}

// AllRoutes collects complete routes in order; no matches yields an empty slice.
func (c *Course[L]) AllRoutes(legend *L, order TraversalOrder) ([]Route, error) {
	routes := []Route{}
	err := c.routes(legend, order, func(route []any) bool {
		routes = append(routes, route)
		return true
	})
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (c *Course[L]) routes(legend *L, order TraversalOrder, yield func([]any) bool) error {
	if c == nil || c.navigator == nil || len(c.steps) == 0 {
		return errors.New("route search requires a prepared Course")
	}
	if order != LegendOrder && order != ChartOrder {
		return fmt.Errorf("invalid route order %q", order)
	}
	start, err := traversal.Start(c.navigator.grammar, legend)
	if err != nil {
		return err
	}
	return start.Routes(c.steps, order == ChartOrder, yield)
}
