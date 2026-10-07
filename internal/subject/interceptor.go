package subject

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Interceptor registers an ordinary Go interface and its typed callback.
// Construct one with Intercept; a zero Interceptor is invalid.
type Interceptor struct {
	invoke func(reflect.Value, Encounter) (Flow, error)
	err    error
}

// Interceptors offers each occurrence to registrations in order, until one
// accepts it. Unmatched contracts and Pass continue to the next registration.
type Interceptors []Interceptor

// Intercept binds the original live carrier to I before calling intercept.
// I must be an ordinary Go interface; Self templates are not specialized.
// Addressable concrete values prefer their pointer receiver. Dynamic interface
// values retain ordinary Go value semantics; Sail never copies a value to
// manufacture pointer methods. Invalid registrations are reported by Propagate.
func Intercept[I any](intercept func(I, Encounter) (Flow, error)) Interceptor {
	contract := reflect.TypeFor[I]()
	if contract.Kind() != reflect.Interface {
		return Interceptor{err: fmt.Errorf("Intercept[%v] requires an interface contract", contract)}
	}
	if intercept == nil {
		return Interceptor{err: fmt.Errorf("Intercept[%v] requires a nonnil callback", contract)}
	}
	return Interceptor{invoke: func(value reflect.Value, at Encounter) (Flow, error) {
		receiver, present, err := reflectutils.BindInterface[I](value)
		if err != nil || !present {
			return Pass, err
		}
		return intercept(receiver, at)
	}}
}

func (interceptors Interceptors) validate() error {
	for index, interceptor := range interceptors {
		if interceptor.err != nil {
			return fmt.Errorf("interceptor %d: %w", index+1, interceptor.err)
		}
		if interceptor.invoke == nil {
			return fmt.Errorf("interceptor %d requires a registration", index+1)
		}
	}
	return nil
}

func (interceptors Interceptors) intercept(navigator *engine.SubjectNavigator, at Encounter) (Flow, error) {
	value, present, err := navigator.Carrier(at.occurrence)
	if err != nil || !present {
		return Prune, err
	}
	for index, interceptor := range interceptors {
		// A preceding Pass may have replaced or cleared this carrier.
		if index > 0 {
			value, present, err = navigator.Carrier(at.occurrence)
			if err != nil || !present {
				return Prune, err
			}
		}
		flow, err := interceptor.invoke(value, at)
		if err != nil {
			return Pass, fmt.Errorf("interceptor %d: %w", index+1, err)
		}
		switch flow {
		case Pass:
			continue
		case Descend, Prune, Stop:
			return flow, nil
		default:
			return Pass, fmt.Errorf("interceptor %d returned invalid flow %d", index+1, flow)
		}
	}
	return Descend, nil
}
