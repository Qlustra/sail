package subject

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/reflectutils"
)

// Interceptor registers an ordinary Go interface and its typed callback.
// Construct one with Intercept; a zero Interceptor is invalid.
type Interceptor struct {
	invoke func(reflect.Value, Encounter) (Route, error)
	err    error
}

// Intercept binds the original live carrier to I before calling intercept.
// I must be an ordinary Go interface. Addressable values prefer their pointer
// receiver; yielded values retain Go value semantics without invented addresses.
// Invalid registrations are reported before propagation invokes any callbacks.
func Intercept[I any](intercept func(I, Encounter) (Route, error)) Interceptor {
	contract := reflect.TypeFor[I]()
	if contract.Kind() != reflect.Interface {
		return Interceptor{err: fmt.Errorf("Intercept[%v] requires an interface contract", contract)}
	}
	if intercept == nil {
		return Interceptor{err: fmt.Errorf("Intercept[%v] requires a nonnil callback", contract)}
	}
	return Interceptor{invoke: func(value reflect.Value, at Encounter) (Route, error) {
		receiver, present, err := reflectutils.BindInterface[I](value)
		if err != nil || !present {
			return Pass(), err
		}
		return intercept(receiver, at)
	}}
}

func (registrations Propagation) validate() error {
	for index, interceptor := range registrations {
		if interceptor.err != nil {
			return fmt.Errorf("interceptor %d: %w", index+1, interceptor.err)
		}
		if interceptor.invoke == nil {
			return fmt.Errorf("interceptor %d requires a registration", index+1)
		}
	}
	return nil
}
