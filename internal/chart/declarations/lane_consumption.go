package declarations

import (
	"fmt"
	"strings"

	"github.com/qlustra/intel/logic"
)

// AnalyzeLaneConsumption checks selector-level reachability under committed
// eager consumption. It does not attempt to prove arbitrary Chart type rules
// satisfiable. Results depend only on the completed grammar, never live values.
func AnalyzeLaneConsumption(legend *Legend) ([]Warning, logic.Check) {
	if legend == nil {
		return nil, logic.Nope("invalid_lane_consumption", "legend declaration is nil")
	}
	var warnings []Warning
	check := walkBodies(legend, func(path string, uses []*LaneCarrier) logic.Check {
		for index, current := range uses {
			if current == nil || current.Declaration == nil || current.Exploration != EagerExploration {
				continue
			}
			// Leaving current requires a selector it does not claim. Such a
			// boundary can occur in an intervening optional Lane or in the first
			// required Lane itself. Unknown App fields cannot provide a boundary.
			var boundaries []string
			seenBoundary := make(map[selector]bool)
			for _, later := range uses[index+1:] {
				if later == nil || later.Declaration == nil {
					continue
				}
				var required, shadowedRequired []string
				for _, spot := range later.Declaration.spots {
					key := spotSelector(spot)
					if current.Declaration.selectors[key] == nil {
						if !seenBoundary[key] {
							seenBoundary[key] = true
							label := fmt.Sprintf("family=%q", key.family)
							if key.discriminators != "" {
								label += fmt.Sprintf(" discriminate=%q", key.discriminators)
							}
							// Describe selectors, not claiming Spots: an earlier
							// intervening Lane may own the same boundary selector.
							boundaries = append(boundaries, label)
						}
					} else if spot.Required {
						shadowedRequired = append(shadowedRequired, spotField(spot))
					}
					if spot.Required {
						required = append(required, spotField(spot))
					}
				}
				if len(required) == 0 {
					continue
				}
				if len(boundaries) == 0 {
					return logic.Nope("unreachable_lane_requirement", fmt.Sprintf("%s: required spots %s in lane %q are unreachable: eager lane %q claims every selector that could enter it, including intervening optional lanes", path, strings.Join(required, ", "), laneField(later), laneField(current)))
				}
				if len(shadowedRequired) != 0 {
					warnings = append(warnings, Warning{
						Code:    "eager_lane_entry_boundary",
						Message: fmt.Sprintf("%s: eager lane %q claims selectors needed by required spots %s in lane %q; reaching those occurrences requires first leaving it using one of these boundary selectors: %s", path, laneField(current), strings.Join(shadowedRequired, ", "), laneField(later), strings.Join(boundaries, "; ")),
					})
				}
				// Any successful Chart must have left current before satisfying
				// this required Lane. Later requirements are not shadowed by current.
				break
			}
		}
		return logic.Ok()
	}, nil)
	if !check.OK() {
		return nil, check // Failed compilation never publishes partial advisories.
	}
	return warnings, logic.Ok()
}
