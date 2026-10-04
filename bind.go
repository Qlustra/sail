package sail

import (
	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Bind resolves a populated marker against a live application and returns its
// receiver as I. The application must be a nonnil pointer to the recorded root;
// additional pointer wrappers are allowed. I must be an ordinary Go interface;
// Self templates are not specialized during binding.
//
// Binding follows only the recorded field path, using existing storage. It
// never allocates missing application values, constructs surrogates, or invokes
// application methods. Addressable concrete fields bind through their address when possible.
// Interface fields expose their current dynamic value, with ordinary Go value
// semantics when that value is not a pointer. Rebind after replacing a pointer
// or interface field to obtain the new receiver.
//
// An existing receiver without I returns present=false with no error. Invalid
// markers or roots, nil path/receiver storage, inaccessible pointer receivers,
// and cyclic receiver resolution return errors. Nil maps and slices are values,
// not missing storage. Method invocation and its errors belong to the caller.
func Bind[I any](app any, marker legend.MarkerDeclaration) (receiver I, present bool, err error) {
	return engine.Bind[I](app, marker)
}
