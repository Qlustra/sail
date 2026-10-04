package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// acceptance retains the declaration and any compiled method signatures.
// Its zero value represents a descriptor without an Accept hook.
type acceptance struct {
	declaration Acceptance
	contract    capabilityContract
}

func (d *markerDescriptor) collectAcceptance(surrogate any) error {
	provider, ok := surrogate.(AcceptanceProvider)
	if !ok {
		return nil
	}
	declaration, err := utilities.InvokeDeclaration(d.typ, "Accept", provider.Accept)
	if err != nil {
		return err
	}
	d.acceptance, err = compileAcceptance(declaration)
	if err != nil {
		return fmt.Errorf("%v.Accept(): %w", d.typ, err)
	}
	return nil
}

func compileAcceptance(declaration Acceptance) (acceptance, error) {
	compiled := acceptance{declaration: declaration}
	switch declaration.mode {
	case acceptExact:
		// AcceptExact captures a valid type directly from its generic argument.
	case acceptKind:
		if declaration.kind <= reflect.Invalid || declaration.kind > reflect.UnsafePointer {
			return acceptance{}, fmt.Errorf("invalid acceptance kind %v", declaration.kind)
		}
	case acceptContract:
		contract, err := compileCapabilityContract(declaration.typ)
		if err != nil {
			return acceptance{}, err
		}
		compiled.contract = contract
	default:
		return acceptance{}, fmt.Errorf("returned an empty acceptance declaration")
	}
	return compiled, nil
}

func (a acceptance) match(receivingType reflect.Type) error {
	switch a.declaration.mode {
	case acceptExact:
		if receivingType != a.declaration.typ {
			return fmt.Errorf("expected exactly %v, got %v", a.declaration.typ, receivingType)
		}
	case acceptKind:
		if receivingType.Kind() != a.declaration.kind {
			return fmt.Errorf("expected kind %v, got %v (kind %v)", a.declaration.kind, receivingType, receivingType.Kind())
		}
	case acceptContract:
		typ := utilities.ResolvePointerChain(receivingType)
		if typ == nil {
			return fmt.Errorf("cannot resolve receiving type %v for contract %v", receivingType, a.contract.typ)
		}
		if err := a.contract.match(typ); err != nil {
			return fmt.Errorf("expected contract %v, got %v: %w", a.contract.typ, receivingType, err)
		}
	}
	return nil
}
