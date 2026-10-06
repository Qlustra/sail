package legend

import (
	"reflect"

	"github.com/qlustra/sail/internal/declaration"
)

// Acceptance declares a constraint on a marker's receiving type. Its contents
// are immutable; returning the zero value from Accept is a compilation error.
type Acceptance = declaration.Acceptance

// AcceptExact accepts only the original declared type T. Aliases retain identity;
// distinct defined types and pointers are not normalized. An interface T means
// that exact interface declaration, not any type implementing it.
func AcceptExact[T any]() Acceptance { return declaration.AcceptExact[T]() }

// AcceptKind accepts declarations of kind, including defined types with that
// kind. It does not dereference pointers or constrain collection elements.
// Invalid and unknown kinds fail compilation rather than panicking here.
func AcceptKind(kind reflect.Kind) Acceptance { return declaration.AcceptKind(kind) }

// AcceptContract requires the interface contract I without invoking its methods.
// Contract matching resolves Application pointer chains, retaining defined-type
// identity. Concrete T supports methods on T or *T; interface declarations use
// their own method set. Self refers to that resolved receiving type.
// Invalid contracts fail compilation rather than panicking here.
func AcceptContract[I any]() Acceptance { return declaration.AcceptContract[I]() }

// AcceptanceProvider optionally constrains the Application type for a Legend,
// Region, or Landmark. Accept follows the zero-value contract of Mark and runs
// once per marker declaration type during compilation. The constraint supplements
// the marker's role requirements and never participates in mark selection.
type AcceptanceProvider = declaration.AcceptanceProvider
