package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Flow determines how propagation proceeds after an interceptor is called.
type Flow = subjectcore.Flow

const (
	// Pass declines this interceptor and tries the next registration.
	Pass = subjectcore.Pass
	// Descend accepts the occurrence and visits its declared children.
	Descend = subjectcore.Descend
	// Prune accepts the occurrence, skips its children, and continues siblings.
	Prune = subjectcore.Prune
	// Stop finishes the entire propagation successfully.
	Stop = subjectcore.Stop
)
