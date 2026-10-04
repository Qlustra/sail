package legend

import "github.com/qlustra/sail/internal/declaration"

// Lane groups unordered spot alternatives into an ordered declaration band.
type Lane = declaration.Lane

type LaneExplorationStrategy = declaration.LaneExplorationStrategy

const (
	// ExploreEagerly stays in the current lane while it accepts a field.
	ExploreEagerly = declaration.ExploreEagerly
	// ExploreUnambiguously requires structurally distinct lane choices. It is
	// the default when the exploration tag is omitted.
	ExploreUnambiguously = declaration.ExploreUnambiguously
)
