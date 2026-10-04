package engine

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/reflectutils"
)

// allocateDeclaration only initializes pointers along a selected declaration path.
// Application types are inspected, never allocated here.
func allocateDeclaration(value reflect.Value) (reflect.Value, error) {
	if reflectutils.ResolvePointerChain(value.Type()) == nil {
		return reflect.Value{}, fmt.Errorf("%v: cyclic declaration pointer type", value.Type())
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			if !value.CanSet() {
				return reflect.Value{}, fmt.Errorf("%v: declaration pointer is not settable", value.Type())
			}
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	return value, nil
}

func declarationField(value reflect.Value, path []int) (reflect.Value, error) {
	for _, index := range path {
		var err error
		value, err = allocateDeclaration(value)
		if err != nil {
			return reflect.Value{}, err
		}
		value = value.Field(index)
	}
	if !value.CanSet() || !value.CanAddr() || !value.Addr().CanInterface() {
		return reflect.Value{}, fmt.Errorf("%v: declaration field is not accessible", value.Type())
	}
	return value, nil
}

func declarationCarrier(value reflect.Value, path []int) (reflect.Value, error) {
	field, err := declarationField(value, path)
	if err != nil {
		return reflect.Value{}, err
	}
	field, err = allocateDeclaration(field)
	if err != nil {
		return reflect.Value{}, err
	}
	return declarationField(field, nil)
}
