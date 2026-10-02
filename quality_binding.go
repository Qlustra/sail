package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

type quality struct {
	name        string
	schemaField []int
	value       reflect.Type
	contract    qualityContract
	method      reflect.Method
}

func compileQuality(field reflect.StructField) (quality, error) {
	typ := utilities.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !utilities.TypeImplements[qualityDeclaration](typ) {
		return quality{}, fmt.Errorf("quality must embed sail.Quality")
	}
	provider := reflect.New(typ).Interface().(qualityDeclaration)
	shape, err := utilities.InvokeDeclaration(typ, "quality", provider.quality)
	if err != nil {
		return quality{}, err
	}
	contract, err := compileQualityContract(shape.capability)
	if err != nil {
		return quality{}, err
	}
	name, err := readSingleTag(field.Tag, "populate", validateSymbol)
	if err != nil {
		return quality{}, err
	}
	method, err := shape.method(name)
	if err != nil {
		return quality{}, err
	}
	path, err := carrierPath(typ, shape.carrier)
	if err != nil {
		return quality{}, err
	}
	return quality{
		name:        field.Name,
		schemaField: slices.Concat(field.Index, path),
		value:       shape.value,
		contract:    contract,
		method:      method,
	}, nil
}

func (shape qualityShape) method(name string) (reflect.Method, error) {
	if containsSelf(shape.value, make(map[reflect.Type]bool)) {
		return reflect.Method{}, fmt.Errorf("quality value type %v contains Self; use any to store a subject-dependent result", shape.value)
	}
	method, exists := shape.capability.MethodByName(name)
	if !exists || !method.IsExported() {
		return reflect.Method{}, fmt.Errorf("quality capability %v must declare exported method %q", shape.capability, name)
	}
	if method.Type.NumIn() != 0 || method.Type.NumOut() != 1 {
		return reflect.Method{}, fmt.Errorf("%v.%s must take no arguments and return exactly one value", shape.capability, name)
	}
	result := method.Type.Out(0)
	// A Self result can only be checked once the receiving type is known.
	if result != reflect.TypeFor[Self]() && !result.AssignableTo(shape.value) {
		return reflect.Method{}, fmt.Errorf("%v.%s returns %v, which is not assignable to quality value %v", shape.capability, name, result, shape.value)
	}
	return method, nil
}

func (q quality) populate(value reflect.Value, subject reflect.Type) error {
	typ := utilities.ResolvePointerChain(subject)
	if typ == nil || typ.Kind() == reflect.Interface {
		return fmt.Errorf("required quality: %v cannot supply a declaration surrogate for %v", subject, q.contract.typ)
	}
	if err := q.contract.match(typ); err != nil {
		return err
	}
	resultType := resolveSelf(q.method.Type.Out(0), typ)
	if !resultType.AssignableTo(q.value) {
		return fmt.Errorf("%v.%s returns %v, which is not assignable to quality value %v", typ, q.method.Name, resultType, q.value)
	}
	destination, err := declarationCarrier(value, q.schemaField)
	if err != nil {
		return err
	}
	method := reflect.New(typ).MethodByName(q.method.Name)
	result, err := utilities.InvokeDeclaration(typ, q.method.Name, func() reflect.Value {
		return method.Call(nil)[0]
	})
	if err != nil {
		return err
	}
	destination.Addr().Interface().(qualityDeclaration).fillQuality(result)
	return nil
}
