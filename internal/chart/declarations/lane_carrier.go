package declarations

import (
	"fmt"

	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/logic"
)

type ExplorationPolicy string

const (
	EagerExploration       ExplorationPolicy = "eager"
	UnambiguousExploration ExplorationPolicy = "unambiguous"
)

// LaneCarrier owns metadata for one field use, not for the shared Lane body
// produced by body compilation.
type LaneCarrier struct {
	*define.Carrier[Lane]
	ID          string // Empty for an unnamed use; otherwise unique within the enclosing body.
	Exploration ExplorationPolicy
}

type explorationTag struct{ inspect.TagFormat }

func (t explorationTag) Validate() logic.Check {
	if ExplorationPolicy(t.Raw) != EagerExploration && ExplorationPolicy(t.Raw) != UnambiguousExploration {
		return logic.Nope("invalid_exploration", fmt.Sprintf("invalid exploration policy %q", t.Raw))
	}
	return logic.Ok()
}

func NewLaneCarrier(carrier *define.Carrier[Lane]) (*LaneCarrier, logic.Check) {
	if carrier == nil || carrier.Property == nil {
		return nil, logic.Nope("invalid_lane_carrier", "lane carrier requires a property")
	}
	policy := UnambiguousExploration
	tags := carrier.Property.Tags()
	id, named := tags.Lookup("lane")
	if named && id == "" {
		return nil, logic.Nope("invalid_lane_id", fmt.Sprintf("field %q: empty lane ID", carrier.Property.FieldName()))
	}
	// Omission defaults; an explicitly empty value must be parsed and rejected.
	if _, present := tags.Lookup("exploration"); present {
		tag, check := tags.Parse[explorationTag]("exploration")
		if !check.OK() {
			return nil, logic.NopeBecause("invalid_lane_exploration", fmt.Sprintf("field %q", carrier.Property.FieldName()), check)
		}
		policy = ExplorationPolicy(tag.Format.Raw)
	}
	// Cross-lane validity is checked on the completed graph, not here.
	return &LaneCarrier{Carrier: carrier, ID: id, Exploration: policy}, logic.Ok()
}

func addLane(uses *[]*LaneCarrier, lane *LaneCarrier) logic.Check {
	if lane.ID != "" {
		for _, previous := range *uses {
			if previous.ID == lane.ID {
				return logic.Nope("duplicate_lane_id", fmt.Sprintf("fields %q and %q use the same lane ID %q", previous.Property.FieldName(), lane.Property.FieldName(), lane.ID))
			}
		}
	}
	*uses = append(*uses, lane)
	return logic.Ok()
}
