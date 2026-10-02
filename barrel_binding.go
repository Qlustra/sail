package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

// A barrel describes one destination and its interface method. Repeated
// bindings remain separate destinations and invoke independently.
type barrel struct {
	name        string
	schemaField []int
	capability  reflect.Type
	method      reflect.Method
	required    bool
}

func (l *location) collectBarrel(field reflect.StructField) error {
	declaration, err := compileBarrel(field)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", l.typ, field.Name, err)
	}
	l.barrels = append(l.barrels, declaration)
	return nil
}

func compileBarrel(field reflect.StructField) (barrel, error) {
	typ := utilities.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !utilities.TypeImplements[barrelDeclaration](typ) {
		return barrel{}, fmt.Errorf("barrel must embed sail.Barrel")
	}
	provider := reflect.New(typ).Interface().(barrelDeclaration)
	shape, err := utilities.InvokeDeclaration(typ, "barrel", provider.barrel)
	if err != nil {
		return barrel{}, err
	}
	name, err := readSingleTag(field.Tag, "populate", validateSymbol)
	if err != nil {
		return barrel{}, err
	}
	method, err := shape.method(name)
	if err != nil {
		return barrel{}, err
	}
	required, err := readBarrelRequirement(field.Tag)
	if err != nil {
		return barrel{}, err
	}
	path, err := carrierPath(typ, shape.carrier)
	if err != nil {
		return barrel{}, err
	}
	return barrel{
		name:        field.Name,
		schemaField: slices.Concat(field.Index, path),
		capability:  shape.capability,
		method:      method,
		required:    required,
	}, nil
}

func (shape barrelShape) method(name string) (reflect.Method, error) {
	if shape.capability.Kind() != reflect.Interface {
		return reflect.Method{}, fmt.Errorf("barrel capability %v must be an interface", shape.capability)
	}
	method, exists := shape.capability.MethodByName(name)
	if !exists || !method.IsExported() {
		return reflect.Method{}, fmt.Errorf("barrel capability %v must declare exported method %q", shape.capability, name)
	}
	// Interface method signatures exclude the receiver.
	if method.Type.NumIn() != 0 || method.Type.NumOut() != 1 {
		return reflect.Method{}, fmt.Errorf("%v.%s must take no arguments and return exactly one value", shape.capability, name)
	}
	if !method.Type.Out(0).AssignableTo(shape.value) {
		return reflect.Method{}, fmt.Errorf("%v.%s returns %v, which is not assignable to barrel value %v", shape.capability, name, method.Type.Out(0), shape.value)
	}
	return method, nil
}

func readBarrelRequirement(tags reflect.StructTag) (bool, error) {
	format := SingleValueTag{Tag: Tag{Name: "barrel"}}
	present, err := readTag(tags, "barrel", &format)
	if err != nil {
		return false, err
	}
	if !present || format.Raw == "optional" {
		return false, nil
	}
	if format.Raw == "required" {
		return true, nil
	}
	return false, fmt.Errorf("invalid barrel requirement %q: expected optional or required", format.Raw)
}

func (l *location) populateBarrels(value reflect.Value, subject reflect.Type) error {
	for _, barrel := range l.barrels {
		if err := barrel.populate(value, subject); err != nil {
			return fmt.Errorf("%v.%s: %w", l.typ, barrel.name, err)
		}
	}
	return nil
}

func (b barrel) populate(value reflect.Value, subject reflect.Type) error {
	destination, err := declarationCarrier(value, b.schemaField)
	if err != nil {
		return err
	}
	typ := utilities.ResolvePointerChain(subject)
	// An interface declaration has no concrete implementation to instantiate.
	// Pointer cycles likewise cannot supply a declaration surrogate.
	if typ == nil || typ.Kind() == reflect.Interface || !reflect.PointerTo(typ).Implements(b.capability) {
		if b.required {
			return fmt.Errorf("required barrel: %v cannot supply a declaration surrogate for %v.%s", subject, b.capability, b.method.Name)
		}
		return nil
	}

	// Invoke through the declared interface's method set, whose index was
	// compiled during exploration. No concrete method-name lookup is needed.
	capability := reflect.New(b.capability).Elem()
	capability.Set(reflect.New(typ))
	method := capability.Method(b.method.Index)
	result, err := utilities.InvokeDeclaration(typ, b.method.Name, func() reflect.Value {
		return method.Call(nil)[0]
	})
	if err != nil {
		return err
	}
	destination.Addr().Interface().(barrelDeclaration).fillBarrel(result)
	return nil
}
