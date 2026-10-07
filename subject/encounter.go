package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Encounter describes a live Subject occurrence through its populated Chart
// marker, parent, depth, and carrier reference. The context may be retained
// while the Chart topology remains stable.
type Encounter = subjectcore.Encounter
