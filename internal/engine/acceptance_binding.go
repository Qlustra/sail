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
		contract, err := capability.CompileAcceptanceContract(rule.Type)
		if err != nil {
			return acceptance{}, err
		}
		compiled.contract = contract
	case decl.FamilyContract:
		if rule.Family.Kind() != reflect.Struct {
			return acceptance{}, fmt.Errorf("contract family %v must be a concrete non-pointer struct", rule.Family)
		}
		provider, err := capability.CompileNativeContract(rule.Type)
		if err != nil {
			return acceptance{}, fmt.Errorf("contract family %v provider: %w", rule.Family, err)
		}
		compiled.contract = provider
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
	case decl.FamilyContract:
		return a.matchFamily(carrierType)
	}
	return nil
}

func (a acceptance) matchFamily(carrierType reflect.Type) error {
	typ := reflectutils.ResolvePointerChain(carrierType)
	if typ == nil {
		return fmt.Errorf("cannot resolve carrier type %v for contract family %v", carrierType, a.rule.Family)
	}
	if err := a.contract.Match(typ); err != nil {
		return fmt.Errorf("contract family %v requires provider %v, got %v: %w", a.rule.Family, a.contract.Type(), carrierType, err)
	}
	if typ.Kind() == reflect.Interface {
		return fmt.Errorf("carrier %v cannot supply a concrete zero-valued witness for contract family %v", carrierType, a.rule.Family)
	}

	candidate := reflect.New(typ).Interface()
	contractType, err := capability.InvokeDeclaration(a.rule.Family, "ContractOf", func() reflect.Type {
		return a.rule.Resolve(candidate)
	})
	if err != nil {
		return fmt.Errorf("contract family %v for carrier %v: %w", a.rule.Family, carrierType, err)
	}
	contract, err := capability.CompileNativeContract(contractType)
	if err != nil {
		return fmt.Errorf("contract family %v returned invalid contract for carrier %v: %w", a.rule.Family, carrierType, err)
	}
	if err := contract.Match(typ); err != nil {
		return fmt.Errorf("contract family %v expected contract %v, got %v: %w", a.rule.Family, contractType, carrierType, err)
	}
	return nil
}
