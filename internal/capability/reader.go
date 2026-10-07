package capability

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Reader compiles one declaration method and its destination type. Qualities
// retain it in their descriptors; ReadCapability compiles it for an individual call.
type Reader struct {
	value    reflect.Type
	contract Contract
	method   reflect.Method
}

// CompileReader validates the selected method and result value type.
func CompileReader(value, capability reflect.Type, name string) (Reader, error) {
	contract, err := CompileContract(capability)
	if err != nil {
		return Reader{}, err
	}
	if ContainsSelf(value, make(map[reflect.Type]bool)) {
		return Reader{}, fmt.Errorf("capability value type %v contains Self; use any to store a subject-dependent result", value)
	}
	method, exists := capability.MethodByName(name)
	if !exists || !method.IsExported() {
		return Reader{}, fmt.Errorf("capability %v must declare exported method %q", capability, name)
	}
	if method.Type.NumIn() != 0 || method.Type.NumOut() != 1 {
		return Reader{}, fmt.Errorf("%v.%s must take no arguments and return exactly one value", capability, name)
	}
	result := method.Type.Out(0)
	// A Self result can only be checked once the carrier type is known.
	if result != reflect.TypeFor[declaration.Self]() && !result.AssignableTo(value) {
		return Reader{}, fmt.Errorf("%v.%s returns %v, which is not assignable to capability value %v", capability, name, result, value)
	}
	return Reader{value: value, contract: contract, method: method}, nil
}

// Resolve only reports capability unavailability. Optional readers can skip it;
// required declarations retain the reason in their population error.
func (r Reader) Resolve(carrierType reflect.Type) (reflect.Type, error) {
	typ := reflectutils.ResolvePointerChain(carrierType)
	if typ == nil || typ.Kind() == reflect.Interface {
		return nil, fmt.Errorf("%v cannot supply a declaration surrogate for %v", carrierType, r.contract.typ)
	}
	if err := r.contract.Match(typ); err != nil {
		return nil, err
	}
	return typ, nil
}

// Read requires a resolved carrier type already matched by Resolve. Result type
// mismatches and hook failures are errors even when capability presence is optional.
func (r Reader) Read(carrierType reflect.Type) (reflect.Value, error) {
	resultType := ResolveSelf(r.method.Type.Out(0), carrierType)
	if !resultType.AssignableTo(r.value) {
		return reflect.Value{}, fmt.Errorf("%v.%s returns %v, which is not assignable to capability value %v", carrierType, r.method.Name, resultType, r.value)
	}
	method := reflect.New(carrierType).MethodByName(r.method.Name)
	return InvokeDeclaration(carrierType, r.method.Name, func() reflect.Value {
		return method.Call(nil)[0]
	})
}
