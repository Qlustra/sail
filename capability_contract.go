package sail

import (
	"fmt"
	"reflect"
)

// A capability contract retains the interface as a signature template. Subjects
// need not implement that literal interface when it contains Self.
type capabilityContract struct {
	typ       reflect.Type
	methods   []reflect.Method
	templated bool
}

func compileCapabilityContract(typ reflect.Type) (capabilityContract, error) {
	contract := capabilityContract{typ: typ}
	if typ.Kind() != reflect.Interface {
		return contract, fmt.Errorf("capability %v must be an interface", typ)
	}
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		for _, slot := range signatureTypes(method.Type) {
			if slot == reflect.TypeFor[Self]() {
				contract.templated = true
			} else if containsSelf(slot, make(map[reflect.Type]bool)) {
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

// containsSelf only detects unsupported compound placeholders. It does not
// reconstruct types or specialize Go generic interfaces.
func containsSelf(typ reflect.Type, seen map[reflect.Type]bool) bool {
	if typ == reflect.TypeFor[Self]() {
		return true
	}
	if seen[typ] {
		return false
	}
	seen[typ] = true
	switch typ.Kind() {
	case reflect.Array, reflect.Chan, reflect.Pointer, reflect.Slice:
		return containsSelf(typ.Elem(), seen)
	case reflect.Map:
		return containsSelf(typ.Key(), seen) || containsSelf(typ.Elem(), seen)
	case reflect.Func:
		for _, slot := range signatureTypes(typ) {
			if containsSelf(slot, seen) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			if containsSelf(typ.Field(i).Type, seen) {
				return true
			}
		}
	case reflect.Interface:
		for i := 0; i < typ.NumMethod(); i++ {
			if containsSelf(typ.Method(i).Type, seen) {
				return true
			}
		}
	}
	return false
}

func (c capabilityContract) match(subject reflect.Type) error {
	receiver := reflect.PointerTo(subject)
	if !c.templated {
		if !receiver.Implements(c.typ) {
			return fmt.Errorf("%v requires capability %v", subject, c.typ)
		}
		return nil
	}
	for _, expected := range c.methods {
		actual, exists := receiver.MethodByName(expected.Name)
		if !exists {
			return fmt.Errorf("%v requires capability %v: missing method %s", subject, c.typ, expected.Name)
		}
		if !matchesCapabilitySignature(actual.Type, expected.Type, subject) {
			return fmt.Errorf("%v.%s has signature %v, expected %v with Self = %v", subject, expected.Name, actual.Type, expected.Type, subject)
		}
	}
	return nil
}

func matchesCapabilitySignature(actual, expected, subject reflect.Type) bool {
	// Concrete methods include their receiver; interface methods do not.
	if actual.NumIn() != expected.NumIn()+1 || actual.NumOut() != expected.NumOut() || actual.IsVariadic() != expected.IsVariadic() {
		return false
	}
	for i := 0; i < expected.NumIn(); i++ {
		if actual.In(i+1) != resolveSelf(expected.In(i), subject) {
			return false
		}
	}
	for i := 0; i < expected.NumOut(); i++ {
		if actual.Out(i) != resolveSelf(expected.Out(i), subject) {
			return false
		}
	}
	return true
}

func resolveSelf(typ, subject reflect.Type) reflect.Type {
	if typ == reflect.TypeFor[Self]() {
		return subject
	}
	return typ
}
