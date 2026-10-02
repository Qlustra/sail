package sail

import (
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// ResolvePointerChain returns the first non-pointer type in typ's pointer
// chain, preserving defined-type identity. Nil input and cyclic pointer chains
// return nil. Interface types remain interfaces; resolving them would require
// a live value.
func ResolvePointerChain(typ reflect.Type) reflect.Type {
	return utilities.ResolvePointerChain(typ)
}
