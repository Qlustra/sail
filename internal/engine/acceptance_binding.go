package engine

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/capability"
	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// acceptance retains the declaration and any compiled method signatures.
// Its zero value represents a descriptor without an Accept hook.
type acceptance struct {
	rule     decl.AcceptanceRule
	contract capability.Contract
}

func (d *markerDescriptor) collectAcceptance(surrogate any) error {
	provider, ok := surrogate.(decl.AcceptanceProvider)
	if !ok {
		return nil
	}
	declaration, err := capability.InvokeDeclaration(d.typ, "Accept", provider.Accept)
	if err != nil {
		return err
	}
	d.acceptance, err = compileAcceptance(declaration)
	if err != nil {
		return fmt.Errorf("%v.Accept(): %w", d.typ, err)
	}
	return nil
}

func compileAcceptance(spec decl.Acceptance) (acceptance, error) {
	rule := decl.InspectAcceptance(spec)
	compiled := acceptance{rule: rule}
	switch rule.Mode {
	case decl.Exact:
		// AcceptExact captures a valid type directly from its generic argument.
	case decl.Kind:
		if rule.Kind <= reflect.Invalid || rule.Kind > reflect.UnsafePointer {
			return acceptance{}, fmt.Errorf("invalid acceptance kind %v", rule.Kind)
		}
	case decl.Contract:
		contract, err := capability.CompileContract(rule.Type)
		if err != nil {
			return acceptance{}, err
		}
		compiled.contract = contract
	default:
		return acceptance{}, fmt.Errorf("returned an empty acceptance declaration")
	}
	return compiled, nil
}

func (a acceptance) match(carrierType reflect.Type) error {
	switch a.rule.Mode {
	case decl.Exact:
		if carrierType != a.rule.Type {
			return fmt.Errorf("expected exactly %v, got %v", a.rule.Type, carrierType)
		}
	case decl.Kind:
		if carrierType.Kind() != a.rule.Kind {
			return fmt.Errorf("expected kind %v, got %v (kind %v)", a.rule.Kind, carrierType, carrierType.Kind())
		}
	case decl.Contract:
		typ := reflectutils.ResolvePointerChain(carrierType)
		if typ == nil {
			return fmt.Errorf("cannot resolve carrier type %v for contract %v", carrierType, a.contract.Type())
		}
		if err := a.contract.Match(typ); err != nil {
			return fmt.Errorf("expected contract %v, got %v: %w", a.contract.Type(), carrierType, err)
		}
	}
	return nil
}
