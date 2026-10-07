package subject

// Flow determines how propagation proceeds after an interceptor is called.
type Flow uint8

const (
	// Pass declines this interceptor and tries the next registration.
	Pass Flow = iota
	// Descend accepts the occurrence and visits its declared children.
	Descend
	// Prune accepts the occurrence, skips its children, and continues siblings.
	Prune
	// Stop finishes the entire propagation successfully.
	Stop
)
