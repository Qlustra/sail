package capability

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Contract retains an interface or a concrete type's public method signatures.
// Carrier types need not implement a literal interface when it contains Self.
type Contract struct {
	typ       reflect.Type
	methods   []reflect.Method
	templated bool
}

// CompileContract validates an interface and retains its method signatures.
func CompileContract(typ reflect.Type) (Contract, error) {
	contract := Contract{typ: typ}
	if typ.Kind() != reflect.Interface {
		return contract, fmt.Errorf("capability %v must be an interface", typ)
	}
	for i := 0; i < typ.NumMethod(); i++ {
		if err := contract.addMethod(typ.Method(i)); err != nil {
			return contract, err
		}
	}
	// Reflection exposes only exported concrete methods. Ordinary interfaces
	// still use Go's Implements check, including their private methods.
	if contract.templated {
		for _, method := range contract.methods {
			if !method.IsExported() {
				return contract, fmt.Errorf("capability %v uses Self: method %s must be exported", typ, method.Name)
			}
		}
	}
	return contract, nil
}

// CompileNativeContract accepts only ordinary Go interfaces, whose matching
// uses Implements directly without substituting Self or comparing signatures.
func CompileNativeContract(typ reflect.Type) (Contract, error) {
	if typ == nil {
		return Contract{}, fmt.Errorf("contract type must not be nil")
	}
	contract, err := CompileContract(typ)
	if err != nil {
		return Contract{}, err
	}
	if contract.templated {
		return Contract{}, fmt.Errorf("contract %v must be an ordinary Go interface; Self is unsupported", typ)
	}
	return contract, nil
}

// CompileAcceptanceContract also permits a struct to supply its complete
// exported method set, including pointer-receiver methods. Receiver types do
// not participate in signature matching; parameter and result types do.
func CompileAcceptanceContract(typ reflect.Type) (Contract, error) {
	if typ.Kind() == reflect.Interface {
		return CompileContract(typ)
	}
	contract := Contract{typ: typ}
	resolved := reflectutils.ResolvePointerChain(typ)
	if resolved == nil || resolved.Kind() != reflect.Struct {
		return contract, fmt.Errorf("acceptance contract %v must be an interface or struct", typ)
	}
	receiver := reflect.PointerTo(resolved)
	for i := 0; i < receiver.NumMethod(); i++ {
		method := receiver.Method(i)
		inputs := make([]reflect.Type, method.Type.NumIn()-1)
		for i := range inputs {
			inputs[i] = method.Type.In(i + 1)
		}
		outputs := make([]reflect.Type, method.Type.NumOut())
		for i := range outputs {
			outputs[i] = method.Type.Out(i)
		}
		method.Type = reflect.FuncOf(inputs, outputs, method.Type.IsVariadic())
		if err := contract.addMethod(method); err != nil {
			return contract, err
		}
	}
	return contract, nil
}

func (c *Contract) addMethod(method reflect.Method) error {
	for _, slot := range signatureTypes(method.Type) {
		if slot == reflect.TypeFor[declaration.Self]() {
			c.templated = true
		} else if ContainsSelf(slot, make(map[reflect.Type]bool)) {
			return fmt.Errorf("%v.%s: Self must be a direct parameter or result, not part of %v", c.typ, method.Name, slot)
		}
	}
	c.methods = append(c.methods, method)
	return nil
}

func signatureTypes(signature reflect.Type) []reflect.Type {
	result := make([]reflect.Type, 0, signature.NumIn()+signature.NumOut())
	for i := 0; i < signature.NumIn(); i++ {
		result = append(result, signature.In(i))
	}
	for i := 0; i < signature.NumOut(); i++ {
		result = append(result, signature.Out(i))
	}
	return result
}

// ContainsSelf only detects unsupported compound placeholders. It does not
// reconstruct types or specialize Go generic interfaces.
func ContainsSelf(typ reflect.Type, seen map[reflect.Type]bool) bool {
	if typ == reflect.TypeFor[declaration.Self]() {
		return true
	}
	if seen[typ] {
		return false
	}
	seen[typ] = true
	switch typ.Kind() {
	case reflect.Array, reflect.Chan, reflect.Pointer, reflect.Slice:
		return ContainsSelf(typ.Elem(), seen)
	case reflect.Map:
		return ContainsSelf(typ.Key(), seen) || ContainsSelf(typ.Elem(), seen)
	case reflect.Func:
		for _, slot := range signatureTypes(typ) {
			if ContainsSelf(slot, seen) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			if ContainsSelf(typ.Field(i).Type, seen) {
				return true
			}
		}
	case reflect.Interface:
		for i := 0; i < typ.NumMethod(); i++ {
			if ContainsSelf(typ.Method(i).Type, seen) {
				return true
			}
		}
	}
	return false
}

// Type returns the declared contract template.
func (c Contract) Type() reflect.Type { return c.typ }

// Match checks the carrier type against the compiled contract.
func (c Contract) Match(carrierType reflect.Type) error {
	receiver := carrierType
	receiverInputs := 0
	if carrierType.Kind() != reflect.Interface {
		receiver = reflect.PointerTo(carrierType)
		receiverInputs = 1
	}
	if c.typ.Kind() == reflect.Interface && !c.templated {
		if !receiver.Implements(c.typ) {
			return fmt.Errorf("%v requires capability %v", carrierType, c.typ)
		}
		return nil
	}
	for _, expected := range c.methods {
		actual, exists := receiver.MethodByName(expected.Name)
		if !exists {
			return fmt.Errorf("%v requires capability %v: missing method %s", carrierType, c.typ, expected.Name)
		}
		if !matchesCapabilitySignature(actual.Type, expected.Type, carrierType, receiverInputs) {
			return fmt.Errorf("%v.%s has signature %v, expected %v with Self = %v", carrierType, expected.Name, actual.Type, expected.Type, carrierType)
		}
	}
	return nil
}

func matchesCapabilitySignature(actual, expected, carrierType reflect.Type, receiverInputs int) bool {
	// Concrete methods include their receiver; interface methods do not.
	if actual.NumIn() != expected.NumIn()+receiverInputs || actual.NumOut() != expected.NumOut() || actual.IsVariadic() != expected.IsVariadic() {
		return false
	}
	for i := 0; i < expected.NumIn(); i++ {
		if actual.In(i+receiverInputs) != ResolveSelf(expected.In(i), carrierType) {
			return false
		}
	}
	for i := 0; i < expected.NumOut(); i++ {
		if actual.Out(i) != ResolveSelf(expected.Out(i), carrierType) {
			return false
		}
	}
	return true
}

// ResolveSelf substitutes the carrier type for a direct Self placeholder.
func ResolveSelf(typ, carrierType reflect.Type) reflect.Type {
	if typ == reflect.TypeFor[declaration.Self]() {
		return carrierType
	}
	return typ
}
