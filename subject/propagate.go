package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Propagation offers live Subject occurrences to registrations in order, until
// one accepts them. An empty list descends through every present occurrence using
// the explicit order passed to Vessel.Propagate. Registrations are copied on entry.
type Propagation = subjectcore.Propagation
