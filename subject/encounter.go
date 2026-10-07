package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Encounter describes a live Subject occurrence through its populated Chart
// marker, parent, depth, carrier reference, and arrival identity facts. Entered
// member roots retain their containing encounter as parent. First describes
// arrival history, while Cycle reports a matching ancestor before movement is
// chosen. The context may be retained while the Chart topology remains stable.
type Encounter = subjectcore.Encounter
