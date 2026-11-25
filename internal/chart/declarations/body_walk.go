package declarations

import (
	"fmt"

	"github.com/qlustra/intel/logic"
)

// walkBodies visits hosts and first encounters of bodies in authored order.
// Incoming spots remain use-local; shared and recursive bodies are visited once.
// A host path identifies the first incoming use of a shared body, not every
// possible recursive expansion of that body.
func walkBodies(legend *Legend, host func(string, []*LaneCarrier) logic.Check, body func(*Spot) logic.Check) logic.Check {
	regions := make(map[*Region]bool)
	landmarks := make(map[*Landmark]bool)
	var walk func(string, []*LaneCarrier) logic.Check
	walk = func(path string, uses []*LaneCarrier) logic.Check {
		if host != nil {
			if check := host(path, uses); !check.OK() {
				return check
			}
		}
		for _, use := range uses {
			if use == nil || use.Declaration == nil {
				continue
			}
			for _, spot := range use.Declaration.spots {
				if spot == nil {
					continue
				}
				switch {
				case spot.For == SpotForRegion && spot.Region != nil && spot.Region.Declaration != nil:
					region := spot.Region.Declaration
					if regions[region] {
						continue
					}
					regions[region] = true
					if body != nil {
						if check := body(spot); !check.OK() {
							return check
						}
					}
					if check := walk(path+"."+laneField(use)+"."+spotField(spot), region.Lanes); !check.OK() {
						return check
					}
				case spot.For == SpotForLandmark && spot.Landmark != nil && spot.Landmark.Declaration != nil:
					landmark := spot.Landmark.Declaration
					if landmarks[landmark] {
						continue
					}
					landmarks[landmark] = true
					if body != nil {
						if check := body(spot); !check.OK() {
							return check
						}
					}
				}
			}
		}
		return logic.Ok()
	}
	return walk(fmt.Sprintf("legend %q", legend.Mark), legend.Lanes)
}
