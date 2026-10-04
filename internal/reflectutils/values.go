package reflectutils

import (
	"fmt"
	"reflect"
)

// Dereference follows existing pointer storage without allocating missing values.
// Interfaces remain intact because their dynamic value is not a declared path.
func Dereference(value reflect.Value) (reflect.Value, error) {
	if !value.IsValid() {
		return reflect.Value{}, fmt.Errorf("missing value")
	}
	if ResolvePointerChain(value.Type()) == nil {
		return reflect.Value{}, fmt.Errorf("cyclic pointer type %v", value.Type())
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("missing storage: nil %v", value.Type())
		}
		value = value.Elem()
	}
	return value, nil
}

// FieldByPath follows exported fields through existing pointer storage. The
// final field is returned with its declared type, including pointers/interfaces.
func FieldByPath(value reflect.Value, path []int) (reflect.Value, error) {
	for _, index := range path {
		var err error
		value, err = Dereference(value)
		if err != nil {
			return reflect.Value{}, err
		}
		if value.Kind() != reflect.Struct || index < 0 || index >= value.NumField() {
			return reflect.Value{}, fmt.Errorf("field index %d cannot be resolved in %v", index, value.Type())
		}
		field := value.Type().Field(index)
		if !field.IsExported() {
			return reflect.Value{}, fmt.Errorf("field %v.%s is not exported", value.Type(), field.Name)
		}
		value = value.Field(index)
	}
	if !value.IsValid() || !value.CanInterface() {
		return reflect.Value{}, fmt.Errorf("field storage is inaccessible")
	}
	return value, nil
}

// BindInterface obtains an ordinary Go interface from existing storage. It
// prefers an addressable concrete value's address and never creates a copy to
// supply pointer methods. Interfaces expose their current dynamic receiver.
func BindInterface[I any](value reflect.Value) (receiver I, present bool, err error) {
	contract := reflect.TypeFor[I]()
	if contract.Kind() != reflect.Interface {
		return receiver, false, fmt.Errorf("binding contract %v must be an interface", contract)
	}
	type pointer struct {
		typ     reflect.Type
		address uintptr
	}
	seen := make(map[pointer]bool)
	for {
		if !value.IsValid() || !value.CanInterface() {
			return receiver, false, fmt.Errorf("receiver storage is inaccessible")
		}
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return receiver, false, fmt.Errorf("missing storage: nil %v", value.Type())
			}
			value = value.Elem()
			continue
		case reflect.Pointer:
			if value.IsNil() {
				return receiver, false, fmt.Errorf("missing storage: nil %v", value.Type())
			}
			key := pointer{typ: value.Type(), address: value.Pointer()}
			if seen[key] {
				return receiver, false, fmt.Errorf("cyclic receiver through %v", value.Type())
			}
			seen[key] = true
			if receiver, present = value.Interface().(I); present {
				return receiver, true, nil
			}
			value = value.Elem()
			continue
		}
		if value.CanAddr() && value.Addr().CanInterface() {
			if receiver, present = value.Addr().Interface().(I); present {
				return receiver, true, nil
			}
		}
		if receiver, present = value.Interface().(I); present {
			return receiver, true, nil
		}
		if reflect.PointerTo(value.Type()).Implements(contract) {
			return receiver, false, fmt.Errorf("%v requires addressable storage to bind %v", value.Type(), contract)
		}
		return receiver, false, nil
	}
}
