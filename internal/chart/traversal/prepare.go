package traversal

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/chart/declarations"
)

// WaypointShape contains the static restrictions of a waypoint. An empty LaneID
// permits any incoming Lane use.
type WaypointShape struct {
	Marker reflect.Type
	LaneID string
}

// ValidateWaypoints checks for a possible route through the compiled topology.
// It does not evaluate predicates or establish that a populated route exists.
func ValidateWaypoints(grammar *declarations.Legend, steps []WaypointShape) error {
	if grammar == nil {
		return fmt.Errorf("prepare requires a compiled grammar")
	}
	lanes := grammar.Lanes
	terminal := false
	for i, step := range steps {
		var next []*declarations.LaneCarrier
		seenBodies := make(map[*declarations.Region]bool)
		matched := false
		for _, lane := range lanes {
			if step.LaneID != "" && lane.ID != step.LaneID {
				continue
			}
			for _, spot := range lane.Declaration.Spots() {
				var marker reflect.Type
				if spot.For == declarations.SpotForRegion {
					marker = spot.Region.Property.FieldTypeInspector().CapturedType()
				} else {
					marker = spot.Landmark.Property.FieldTypeInspector().CapturedType()
				}
				if spot.Kind == declarations.MultiSpot {
					marker = marker.Elem()
				}
				// Mirror pointer.(T), not Go assignment: a distinct defined
				// pointer type may be assignable without satisfying an assertion.
				pointer := reflect.PointerTo(marker)
				if pointer != step.Marker && !(step.Marker != nil && step.Marker.Kind() == reflect.Interface && pointer.Implements(step.Marker)) {
					continue
				}
				matched = true
				if spot.For == declarations.SpotForLandmark {
					continue
				}
				body := spot.Region.Declaration
				if seenBodies[body] {
					continue
				}
				seenBodies[body] = true
				next = append(next, body.Lanes...)
			}
		}
		if !matched {
			if terminal {
				return fmt.Errorf("waypoint %d: requested type %v in lane %q cannot continue from a terminal Landmark", i+1, step.Marker, step.LaneID)
			}
			return fmt.Errorf("waypoint %d: no reachable token matches type %v in lane %q", i+1, step.Marker, step.LaneID)
		}
		terminal = len(seenBodies) == 0
		lanes = next
	}
	return nil
}
