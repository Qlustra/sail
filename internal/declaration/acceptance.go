package declaration

import "reflect"

type Acceptance struct {
	mode AcceptanceMode
	typ  reflect.Type
	kind reflect.Kind
}

type AcceptanceMode uint8

const (
	Exact AcceptanceMode = iota + 1
	Kind
	Contract
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

type AcceptanceProvider interface {
	Accept() Acceptance
}

type AcceptanceRule struct {
	Mode AcceptanceMode
	Type reflect.Type
	Kind reflect.Kind
}

func InspectAcceptance(acceptance Acceptance) AcceptanceRule {
	return AcceptanceRule{Mode: acceptance.mode, Type: acceptance.typ, Kind: acceptance.kind}
}
