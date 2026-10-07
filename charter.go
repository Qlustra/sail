package sail

import (
	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Charter holds the reusable descriptors compiled from legend L.
// Populate uses it to interpret a Subject declaration into a fresh Chart.
// Descriptor storage is internal; callers use the Charter through Sail operations.
type Charter[L legend.Declaration] struct {
	compiled *engine.Charter
}
