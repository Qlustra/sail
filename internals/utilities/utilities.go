package utilities

import (
	"fmt"
	"reflect"
)

// ResolvePointerChain returns the first non-pointer type in t's pointer chain.
// It returns nil for nil input or a cycle of named pointer types, which has no
// terminal non-pointer type.
func ResolvePointerChain(t reflect.Type) reflect.Type {
	seen := make(map[reflect.Type]struct{})
	for t != nil && t.Kind() == reflect.Pointer {
		if _, exists := seen[t]; exists {
			return nil
		}
		seen[t] = struct{}{}
		t = t.Elem()
	}
	return t
}

func FieldImplements[T any](f reflect.StructField) bool {
	return TypeImplements[T](f.Type)
}

// TypeImplements reports whether t or, for a non-pointer type, *t implements I.
// I must be an interface type.
func TypeImplements[I any](t reflect.Type) bool {
	contract := reflect.TypeFor[I]()
	if contract.Kind() != reflect.Interface {
		panic("implementation target must be an interface type")
	}
	if t == nil {
		return false
	}
	if t.Implements(contract) {
		return true
	}
	return t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(contract)
}

// InvokeDeclaration confines recovery to declaration metadata callbacks. They
// run on newly allocated zero-valued declarations, so promoted methods through
// nil embedded pointers must also be safe for that zero value. Reflection and
// compiler operations remain outside this boundary.
func InvokeDeclaration[T any](typ reflect.Type, method string, invoke func() T) (result T, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			err = fmt.Errorf("%v.%s(): declaration metadata method panicked on its zero-valued receiver: %v", typ, method, failure)
		}
	}()
	return invoke(), nil
}
