package declaration

import "reflect"

type Acceptance struct {
	mode AcceptanceMode
	typ  reflect.Type
	kind reflect.Kind

	family  reflect.Type
	resolve func(any) reflect.Type
}

type AcceptanceMode uint8

const (
	Exact AcceptanceMode = iota + 1
	Kind
	Contract
	FamilyContract
)

func AcceptExact[T any]() Acceptance {
	return Acceptance{mode: Exact, typ: reflect.TypeFor[T]()}
}

func AcceptKind(kind reflect.Kind) Acceptance {
	return Acceptance{mode: Kind, kind: kind}
}

func AcceptContract[T any]() Acceptance {
	return Acceptance{mode: Contract, typ: reflect.TypeFor[T]()}
}

type ContractFamily[Carrier any] interface {
	ContractOf(Carrier) reflect.Type
}

func AcceptContractFamily[Family ContractFamily[Carrier], Carrier any]() Acceptance {
	return Acceptance{
		mode:   FamilyContract,
		typ:    reflect.TypeFor[Carrier](),
		family: reflect.TypeFor[Family](),
		resolve: func(candidate any) reflect.Type {
			var family Family
			return family.ContractOf(candidate.(Carrier))
		},
	}
}

type AcceptanceProvider interface {
	Accept() Acceptance
}

type AcceptanceRule struct {
	Mode AcceptanceMode
	Type reflect.Type
	Kind reflect.Kind

	Family  reflect.Type
	Resolve func(any) reflect.Type
}

func InspectAcceptance(acceptance Acceptance) AcceptanceRule {
	return AcceptanceRule{
		Mode:    acceptance.mode,
		Type:    acceptance.typ,
		Kind:    acceptance.kind,
		Family:  acceptance.family,
		Resolve: acceptance.resolve,
	}
}
