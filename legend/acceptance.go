package legend

import (
	"reflect"

	"github.com/qlustra/sail/internal/declaration"
)

// Acceptance declares a constraint on a marker's carrier type. Its contents
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

// AcceptContract requires an interface or a struct's complete exported method
// set without invoking its methods. A struct contract includes both value and
// pointer-receiver methods; matching ignores receivers and requires identical
// parameter and result types, including variadicness. Additional methods are
// permitted. Struct fields and private methods are not part of the contract.
// Contract matching resolves pointer chains, retaining defined-type identity.
// Concrete carriers support methods on T or *T; interface declarations use
// their own method set. Self refers to the resolved carrier type.
// Invalid contracts fail compilation rather than panicking here.
func AcceptContract[T any]() Acceptance { return declaration.AcceptContract[T]() }

// ContractFamily selects a carrier's instantiated native interface contract.
// Family descriptors must be concrete non-pointer structs whose ContractOf
// method is safe on a zero-valued receiver. Carrier must be an ordinary Go
// interface; it may include package-private witness methods.
type ContractFamily[Carrier any] = declaration.ContractFamily[Carrier]

// AcceptContractFamily requires the carrier provider interface, resolves its
// full contract through a zero-valued Family descriptor and a fresh zero-valued
// concrete carrier, then checks that contract with Go's implementation rules.
// Carrier is inferred from Family's ContractOf parameter when omitted.
//
// ContractOf must return an ordinary Go interface type and must not depend on
// initialized state. Interface-typed carriers cannot supply a concrete witness.
// Self templates are unsupported here; use an instantiated generic interface.
// This operation never inspects live Subject values or initializes their fields.
func AcceptContractFamily[Family ContractFamily[Carrier], Carrier any]() Acceptance {
	return declaration.AcceptContractFamily[Family, Carrier]()
}

// AcceptanceProvider optionally constrains the carrier type for a Legend,
// Region, or Landmark. Accept follows the zero-value contract of Mark and runs
// once per marker declaration type during compilation. The constraint supplements
// the marker's role requirements and never participates in mark selection.
type AcceptanceProvider = declaration.AcceptanceProvider
