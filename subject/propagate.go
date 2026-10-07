package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Propagation offers live Subject occurrences to ordered interceptors. An empty
// list descends through every present occurrence without invoking callbacks.
type Propagation = subjectcore.Propagation
