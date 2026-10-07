package chart

import chartcore "github.com/qlustra/sail/internal/chart"

// Encounter describes one selected arrival in a Chart. Its parent chain follows
// marker relationships, with lanes and spots recorded as incoming edge context.
// The context may be retained while the Chart's topology remains stable.
type Encounter = chartcore.Encounter
