package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// Acceptance declares a constraint on a marker's receiving type. Its contents
// are immutable; returning the zero value from Accept is an exploration error.
type Acceptance struct {
	exact reflect.Type
}

// Exact accepts only the original declared type T. Aliases retain identity;
// distinct defined types and pointers are not normalized. An interface T means
// that exact interface declaration, not any type implementing it.
func Exact[T any]() Acceptance {
	return Acceptance{exact: reflect.TypeFor[T]()}
}

// AcceptanceProvider optionally constrains the subject type for a chart,
// region, or landmark. Accept follows the same zero-value contract as Mark and
// is invoked once per location during exploration. The constraint supplements
// the marker's role requirements and never participates in mark selection.
type AcceptanceProvider interface {
	Accept() Acceptance
}

func (l *location) collectAcceptance(surrogate any) error {
	provider, ok := surrogate.(AcceptanceProvider)
	if !ok {
		return nil
	}
	declaration, err := utilities.InvokeDeclaration(l.typ, "Accept", provider.Accept)
	if err != nil {
		return err
	}
	if declaration.exact == nil {
		return fmt.Errorf("%v.Accept(): returned an empty acceptance declaration", l.typ)
	}
	l.acceptance = declaration
	return nil
}

func (a Acceptance) match(subject reflect.Type) error {
	// A location without an Accept hook retains the unconstrained zero value.
	// An explicitly returned zero value has already failed exploration.
	if a.exact != nil && subject != a.exact {
		return fmt.Errorf("expected exactly %v, got %v", a.exact, subject)
	}
	return nil
}
