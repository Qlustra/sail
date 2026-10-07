package capability

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/declaration"
)

// Contract retains the interface as a signature template. Subject types
// need not implement that literal interface when it contains Self.
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
		method := typ.Method(i)
		for _, slot := range signatureTypes(method.Type) {
			if slot == reflect.TypeFor[declaration.Self]() {
				contract.templated = true
			} else if ContainsSelf(slot, make(map[reflect.Type]bool)) {
				return contract, fmt.Errorf("%v.%s: Self must be a direct parameter or result, not part of %v", typ, method.Name, slot)
			}
		}
		contract.methods = append(contract.methods, method)
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

// Type returns the declared capability interface.
func (c Contract) Type() reflect.Type { return c.typ }

// Match checks the receiving type against the compiled contract.
func (c Contract) Match(receivingType reflect.Type) error {
	receiver := receivingType
	receiverInputs := 0
	if receivingType.Kind() != reflect.Interface {
		receiver = reflect.PointerTo(receivingType)
		receiverInputs = 1
	}
	if !c.templated {
		if !receiver.Implements(c.typ) {
			return fmt.Errorf("%v requires capability %v", receivingType, c.typ)
		}
		return nil
	}
	for _, expected := range c.methods {
		actual, exists := receiver.MethodByName(expected.Name)
		if !exists {
			return fmt.Errorf("%v requires capability %v: missing method %s", receivingType, c.typ, expected.Name)
		}
		if !matchesCapabilitySignature(actual.Type, expected.Type, receivingType, receiverInputs) {
			return fmt.Errorf("%v.%s has signature %v, expected %v with Self = %v", receivingType, expected.Name, actual.Type, expected.Type, receivingType)
		}
	}
	return nil
}

func matchesCapabilitySignature(actual, expected, receivingType reflect.Type, receiverInputs int) bool {
	// Concrete methods include their receiver; interface methods do not.
	if actual.NumIn() != expected.NumIn()+receiverInputs || actual.NumOut() != expected.NumOut() || actual.IsVariadic() != expected.IsVariadic() {
		return false
	}
	for i := 0; i < expected.NumIn(); i++ {
		if actual.In(i+receiverInputs) != ResolveSelf(expected.In(i), receivingType) {
			return false
		}
	}
	for i := 0; i < expected.NumOut(); i++ {
		if actual.Out(i) != ResolveSelf(expected.Out(i), receivingType) {
			return false
		}
	}
	return true
}

// ResolveSelf substitutes the receiving type for a direct Self placeholder.
func ResolveSelf(typ, receivingType reflect.Type) reflect.Type {
	if typ == reflect.TypeFor[declaration.Self]() {
		return receivingType
	}
	return typ
}
