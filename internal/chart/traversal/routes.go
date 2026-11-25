package traversal

import "github.com/qlustra/sail/internal/chart/declarations"

// WaypointMatch selects an immediate token and returns its route value.
type WaypointMatch func(int, *declarations.LaneCarrier, any) (any, bool)

// Routes yields independent complete routes in depth-first sibling order.
// Returning false from yield stops the entire search. Route values retain their
// ordinary pointer identity and lifetime; only the route slice is copied.
func (p SpotPosition) Routes(steps []WaypointMatch, chartOrder bool, yield func([]any) bool) error {
	route := make([]any, 0, len(steps))
	var search func(SpotPosition, int) (bool, error)
	search = func(position SpotPosition, step int) (bool, error) {
		if step == len(steps) {
			return yield(append([]any{}, route...)), nil
		}
		matches, err := position.immediateMatches(steps[step], chartOrder)
		if err != nil {
			return false, err
		}
		for next, marker := range matches {
			route = append(route, marker)
			more, err := search(next, step+1)
			route = route[:len(route)-1]
			if err != nil || !more {
				return false, err
			}
		}
		return true, nil
	}
	_, err := search(p, 0)
	return err
}
