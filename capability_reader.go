package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// A reader compiles one declaration method and its destination type. Qualities
// retain it in the schema; ReadCapability compiles it for an individual call.
type capabilityReader struct {
	value    reflect.Type
	contract capabilityContract
	method   reflect.Method
}

func compileCapabilityReader(value, capability reflect.Type, name string) (capabilityReader, error) {
	contract, err := compileCapabilityContract(capability)
	if err != nil {
		return capabilityReader{}, err
	}
	if containsSelf(value, make(map[reflect.Type]bool)) {
		return capabilityReader{}, fmt.Errorf("capability value type %v contains Self; use any to store a subject-dependent result", value)
	}
	method, exists := capability.MethodByName(name)
	if !exists || !method.IsExported() {
		return capabilityReader{}, fmt.Errorf("capability %v must declare exported method %q", capability, name)
	}
	if method.Type.NumIn() != 0 || method.Type.NumOut() != 1 {
		return capabilityReader{}, fmt.Errorf("%v.%s must take no arguments and return exactly one value", capability, name)
	}
	result := method.Type.Out(0)
	// A Self result can only be checked once the receiving type is known.
	if result != reflect.TypeFor[Self]() && !result.AssignableTo(value) {
		return capabilityReader{}, fmt.Errorf("%v.%s returns %v, which is not assignable to capability value %v", capability, name, result, value)
	}
	return capabilityReader{value: value, contract: contract, method: method}, nil
}

// resolve only reports capability unavailability. Optional readers can skip it;
// required declarations retain the reason in their population error.
func (r capabilityReader) resolve(subject reflect.Type) (reflect.Type, error) {
	typ := ResolvePointerChain(subject)
	if typ == nil || typ.Kind() == reflect.Interface {
		return nil, fmt.Errorf("%v cannot supply a declaration surrogate for %v", subject, r.contract.typ)
	}
	if err := r.contract.match(typ); err != nil {
		return nil, err
	}
	return typ, nil
}

// read requires a receiving type already matched by resolve. Storage and hook
// failures are errors even when capability presence is optional.
func (r capabilityReader) read(receiver reflect.Type) (reflect.Value, error) {
	resultType := resolveSelf(r.method.Type.Out(0), receiver)
	if !resultType.AssignableTo(r.value) {
		return reflect.Value{}, fmt.Errorf("%v.%s returns %v, which is not assignable to capability value %v", receiver, r.method.Name, resultType, r.value)
	}
	method := reflect.New(receiver).MethodByName(r.method.Name)
	return utilities.InvokeDeclaration(receiver, r.method.Name, func() reflect.Value {
		return method.Call(nil)[0]
	})
}
