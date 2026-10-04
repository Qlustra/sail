package capability

import (
	"fmt"
	"reflect"
)

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
