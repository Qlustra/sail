package sail

import (
	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Bind resolves a populated marker's carrier in a live Subject and returns its
// receiver as I. The Subject must be a nonnil pointer to the recorded root;
// additional pointer wrappers are allowed. I must be an ordinary Go interface;
// Self templates are not specialized during binding.
//
// Binding follows only the recorded carrier path, using existing Subject values. It
// never allocates missing subject values, constructs surrogates, or invokes
// subject methods. Addressable concrete fields bind through their address when possible.
// Interface fields expose their current dynamic value, with ordinary Go value
// semantics when that value is not a pointer. Rebind after replacing a pointer
// or interface field to obtain the new receiver.
//
// An existing receiver without I returns present=false with no error. Invalid
// markers or roots, nil pointers or interfaces along the carrier path,
// inaccessible pointer receivers, and cyclic receiver resolution return errors.
// Nil maps and slices remain valid values. Method invocation and its errors
// belong to the caller.
func Bind[I any](subject any, marker legend.MarkerDeclaration) (receiver I, present bool, err error) {
	return engine.Bind[I](subject, marker)
}
