package engine

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/capability"
	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
	"github.com/qlustra/sail/internal/tags"
)

// A barrel describes one destination and its interface method. Repeated
// bindings remain separate destinations and invoke independently.
type barrelDescriptor struct {
	name            string
	declarationPath []int
	capability      reflect.Type
	method          reflect.Method
	required        bool
}

func (d *markerDescriptor) collectBarrel(field reflect.StructField) error {
	declaration, err := compileBarrel(field)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", d.typ, field.Name, err)
	}
	d.barrels = append(d.barrels, declaration)
	return nil
}

func compileBarrel(field reflect.StructField) (barrelDescriptor, error) {
	typ := reflectutils.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !reflectutils.TypeImplements[decl.BarrelDeclaration](typ) {
		return barrelDescriptor{}, fmt.Errorf("barrel must embed legend.Barrel")
	}
	provider := reflect.New(typ).Interface().(decl.BarrelDeclaration)
	shape, err := capability.InvokeDeclaration(typ, "barrel", func() decl.BarrelShape { return decl.DescribeBarrel(provider) })
	if err != nil {
		return barrelDescriptor{}, err
	}
	name, err := tags.ReadSingle(field.Tag, "populate", validateSymbol)
	if err != nil {
		return barrelDescriptor{}, err
	}
	method, err := barrelMethod(shape, name)
	if err != nil {
		return barrelDescriptor{}, err
	}
	required, err := readBarrelRequirement(field.Tag)
	if err != nil {
		return barrelDescriptor{}, err
	}
	path, err := reflectutils.EmbeddedPath(typ, shape.Carrier)
	if err != nil {
		return barrelDescriptor{}, err
	}
	return barrelDescriptor{
		name:            field.Name,
		declarationPath: slices.Concat(field.Index, path),
		capability:      shape.Capability,
		method:          method,
		required:        required,
	}, nil
}

func barrelMethod(shape decl.BarrelShape, name string) (reflect.Method, error) {
	if shape.Capability.Kind() != reflect.Interface {
		return reflect.Method{}, fmt.Errorf("barrel capability %v must be an interface", shape.Capability)
	}
	method, exists := shape.Capability.MethodByName(name)
	if !exists || !method.IsExported() {
		return reflect.Method{}, fmt.Errorf("barrel capability %v must declare exported method %q", shape.Capability, name)
	}
	// Interface method signatures exclude the receiver.
	if method.Type.NumIn() != 0 || method.Type.NumOut() != 1 {
		return reflect.Method{}, fmt.Errorf("%v.%s must take no arguments and return exactly one value", shape.Capability, name)
	}
	if !method.Type.Out(0).AssignableTo(shape.Value) {
		return reflect.Method{}, fmt.Errorf("%v.%s returns %v, which is not assignable to barrel value %v", shape.Capability, name, method.Type.Out(0), shape.Value)
	}
	return method, nil
}

func readBarrelRequirement(structTags reflect.StructTag) (bool, error) {
	format := tags.SingleValue{Tag: tags.Tag{Name: "barrel"}}
	present, err := tags.Read(structTags, "barrel", &format)
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

func (d *markerDescriptor) populateBarrels(value reflect.Value, receivingType reflect.Type) error {
	for _, barrel := range d.barrels {
		if err := barrel.populate(value, receivingType); err != nil {
			return fmt.Errorf("%v.%s: %w", d.typ, barrel.name, err)
		}
	}
	return nil
}

func (b barrelDescriptor) populate(value reflect.Value, receivingType reflect.Type) error {
	destination, err := declarationCarrier(value, b.declarationPath)
	if err != nil {
		return err
	}
	typ := reflectutils.ResolvePointerChain(receivingType)
	// An interface declaration has no concrete implementation to instantiate.
	// Pointer cycles likewise cannot supply a declaration surrogate.
	if typ == nil || typ.Kind() == reflect.Interface || !reflect.PointerTo(typ).Implements(b.capability) {
		if b.required {
			return fmt.Errorf("required barrel: %v cannot supply a declaration surrogate for %v.%s", receivingType, b.capability, b.method.Name)
		}
		return nil
	}

	// Invoke through the declared interface's method set, whose index was
	// resolved during chart compilation. No concrete method-name lookup is needed.
	receiver := reflect.New(b.capability).Elem()
	receiver.Set(reflect.New(typ))
	method := receiver.Method(b.method.Index)
	result, err := capability.InvokeDeclaration(typ, b.method.Name, func() reflect.Value {
		return method.Call(nil)[0]
	})
	if err != nil {
		return err
	}
	decl.FillBarrel(destination.Addr().Interface().(decl.BarrelDeclaration), result)
	return nil
}
