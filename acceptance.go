package sail

import "reflect"

// Acceptance declares a constraint on a marker's receiving type. Its contents
// are immutable; returning the zero value from Accept is an exploration error.
type Acceptance struct {
	mode acceptanceMode
	typ  reflect.Type
	kind reflect.Kind
}

type acceptanceMode uint8

const (
	acceptExact acceptanceMode = iota + 1
	acceptKind
	acceptContract
)

// AcceptExact accepts only the original declared type T. Aliases retain identity;
// distinct defined types and pointers are not normalized. An interface T means
// that exact interface declaration, not any type implementing it.
func AcceptExact[T any]() Acceptance {
	return Acceptance{mode: acceptExact, typ: reflect.TypeFor[T]()}
}

// AcceptKind accepts declarations of kind, including defined types with that
// kind. It does not dereference pointers or constrain collection elements.
// Invalid and unknown kinds fail exploration rather than panicking here.
func AcceptKind(kind reflect.Kind) Acceptance {
	return Acceptance{mode: acceptKind, kind: kind}
}

// AcceptContract requires the interface contract I without invoking its methods.
// Contract matching resolves subject pointer chains, retaining defined-type
// identity. Concrete T supports methods on T or *T; interface declarations use
// their own method set. Self refers to that resolved receiving type.
// Invalid contracts fail exploration rather than panicking here.
func AcceptContract[I any]() Acceptance {
	return Acceptance{mode: acceptContract, typ: reflect.TypeFor[I]()}
}

// AcceptanceProvider optionally constrains the subject type for a chart,
// region, or landmark. Accept follows the same zero-value contract as Mark and
// is invoked once per location during exploration. The constraint supplements
// the marker's role requirements and never participates in mark selection.
type AcceptanceProvider interface {
	Accept() Acceptance
}
