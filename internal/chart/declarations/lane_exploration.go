package declarations

import (
	"fmt"

	"github.com/qlustra/intel/logic"
)

// ValidateLaneExploration checks each host's ordered lane uses in a completed,
// compiler-produced graph. Region bodies are visited once to close recursion;
// distinct Lane uses within a host are never deduplicated by body identity.
func ValidateLaneExploration(legend *Legend) logic.Check {
	// AnalyzeLaneConsumption separately checks required reachability under eager
	// policies. Disjointness and reachability are different grammar guarantees.
	if legend == nil {
		return logic.Nope("invalid_lane_exploration", "legend declaration is nil")
	}
	return walkBodies(legend, validateLaneUses, nil)
}

func validateLaneUses(path string, uses []*LaneCarrier) logic.Check {
	for currentIndex, current := range uses {
		if current == nil || current.Exploration != UnambiguousExploration || current.Declaration == nil {
			continue
		}
		for _, later := range uses[currentIndex+1:] {
			if later == nil || later.Declaration == nil {
				continue
			}
			for _, spot := range current.Declaration.spots {
				key := spotSelector(spot)
				if other := later.Declaration.selectors[key]; other != nil {
					return logic.Nope("lane_exploration_overlap", fmt.Sprintf("%s: lane %q spot %q overlaps lane %q spot %q: family=%q discriminate=%q", path, laneField(current), spotField(spot), laneField(later), spotField(other), key.family, key.discriminators))
				}
			}
			if hasRequiredSpot(later.Declaration) {
				break
			}
		}
	}
	return logic.Ok()
}

func hasRequiredSpot(lane *Lane) bool {
	for _, spot := range lane.spots {
		if spot != nil && spot.Required {
			return true
		}
	}
	return false
}

func laneField(lane *LaneCarrier) string {
	if lane != nil && lane.Carrier != nil && lane.Property != nil {
		return lane.Property.FieldName()
	}
	return "<unknown>"
}
