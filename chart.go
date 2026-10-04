package sail

import (
	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Chart holds the reusable descriptors explored with legend L.
// Populate uses it to describe an application declaration in a fresh world.
// Descriptor storage is internal; callers operate on the chart through Sail.
type Chart[L legend.Declaration] struct {
	compiled *engine.Chart
}
