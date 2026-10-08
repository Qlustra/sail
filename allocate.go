package sail

import (
	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Allocate creates missing pointer storage along one populated marker's recorded
// carrier path and receiver chain. Subject must be a nonnil pointer to the
// compatible recorded root; additional pointer wrappers are allowed.
//
// Existing pointers and their aliases are preserved. Allocation follows only
// the selected occurrence: it neither visits unrelated fields nor initializes
// map or slice contents, invokes Subject hooks, or enters collections. Nil
// interfaces require a concrete instance. Nil dynamic pointers whose storage
// is not writable are rejected rather than replaced with manufactured copies.
//
// Invalid roots or paths, inaccessible storage, and cyclic pointer resolution
// return errors. Allocation is prepared before any pointer cells are assigned,
// so an error leaves Subject unchanged. Callers must serialize Subject changes.
// Bind remains a separate operation with no allocation side effects.
func Allocate(subject any, marker legend.MarkerDeclaration) error {
	return engine.Allocate(subject, marker)
}
