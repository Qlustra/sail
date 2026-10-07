package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Identity is an opaque, comparable receiver identity within one propagation.
// Value members identify yielded occurrences, not objects in an external cache.
type Identity = subjectcore.Identity
