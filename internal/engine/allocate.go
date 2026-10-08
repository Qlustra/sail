package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Allocate prepares and publishes missing pointer storage for one occurrence.
func Allocate(subject any, marker decl.MarkerDeclaration) error {
	source, err := markerReference(marker)
	if err != nil {
		return fmt.Errorf("sail: allocate: %w", err)
	}
	if err := validateAllocationPath(source); err != nil {
		return fmt.Errorf("sail: allocate: path %v from %v: %w", source.Path, source.Root, err)
	}

	value := reflect.ValueOf(subject)
	if err := validateAllocationRoot(value, source.Root); err != nil {
		return fmt.Errorf("%s: allocate: %w", source, err)
	}
	allocation := pointerAllocation{}
	for value.Type() != source.Root {
		value, err = allocation.pointer(value)
		if err != nil {
			return fmt.Errorf("%s: allocate: %w", source, err)
		}
		value = value.Elem()
	}
	for _, index := range source.Path {
		value, err = allocation.dereference(value, false)
		if err != nil {
			return fmt.Errorf("%s: allocate: %w", source, err)
		}
		value = value.Field(index)
	}
	if !value.IsValid() || !value.CanInterface() {
		return fmt.Errorf("%s: allocate: carrier is inaccessible", source)
	}
	if value.Type() != source.Type {
		return fmt.Errorf("%s: allocate: carrier type %v does not match recorded type %v", source, value.Type(), source.Type)
	}
	if _, err := allocation.dereference(value, true); err != nil {
		return fmt.Errorf("%s: allocate: %w", source, err)
	}

	for _, assignment := range allocation.assignments {
		assignment.destination.Set(assignment.value)
	}
	return nil
}

// Validate the entire declared path before preparing live storage. Avoid calling
// Reference.String until the path is known to be safe for its field lookup.
func validateAllocationPath(source decl.Reference) error {
	typ := source.Root
	for _, index := range source.Path {
		terminal := reflectutils.ResolvePointerChain(typ)
		if terminal == nil {
			return fmt.Errorf("cyclic pointer type %v", typ)
		}
		if terminal.Kind() != reflect.Struct || index < 0 || index >= terminal.NumField() {
			return fmt.Errorf("field index %d cannot be resolved in %v", index, typ)
		}
		field := terminal.Field(index)
		if !field.IsExported() {
			return fmt.Errorf("field %v.%s is not exported", terminal, field.Name)
		}
		typ = field.Type
	}
	if typ != source.Type {
		return fmt.Errorf("carrier type %v does not match recorded type %v", typ, source.Type)
	}
	if reflectutils.ResolvePointerChain(typ) == nil {
		return fmt.Errorf("cyclic pointer type %v", typ)
	}
	return nil
}

func validateAllocationRoot(value reflect.Value, root reflect.Type) error {
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return fmt.Errorf("subject must be a nonnil pointer to %v", root)
	}
	terminal := reflectutils.ResolvePointerChain(value.Type())
	expected := reflectutils.ResolvePointerChain(root)
	if terminal == nil || expected == nil {
		return fmt.Errorf("subject root has a cyclic pointer type")
	}
	if terminal != expected {
		return fmt.Errorf("subject root %v does not match recorded root %v", value.Type(), root)
	}
	for typ := value.Type(); typ != root; typ = typ.Elem() {
		if typ.Kind() != reflect.Pointer {
			return fmt.Errorf("subject cannot reach recorded root %v", root)
		}
	}
	return nil
}

type pointerAssignment struct {
	destination reflect.Value
	value       reflect.Value
}

type pointerAllocation struct {
	assignments []pointerAssignment
}

func (allocation *pointerAllocation) pointer(value reflect.Value) (reflect.Value, error) {
	if !value.IsValid() || value.Kind() != reflect.Pointer || !value.CanInterface() {
		return reflect.Value{}, fmt.Errorf("pointer storage is inaccessible")
	}
	if !value.IsNil() {
		return value, nil
	}
	if !value.CanSet() {
		return reflect.Value{}, fmt.Errorf("nil pointer of type %v has no writable storage; supply an explicit concrete instance", value.Type())
	}

	fresh := reflect.New(value.Type().Elem())
	if !fresh.Type().AssignableTo(value.Type()) {
		return reflect.Value{}, fmt.Errorf("allocated pointer %v cannot be assigned to %v", fresh.Type(), value.Type())
	}
	// A defined pointer type retains its identity in the staged receiver, even
	// when reflect.New returns its unnamed underlying pointer type.
	storage := reflect.New(value.Type()).Elem()
	storage.Set(fresh)
	allocation.assignments = append(allocation.assignments, pointerAssignment{
		destination: value,
		value:       storage,
	})
	return storage, nil
}

func (allocation *pointerAllocation) dereference(value reflect.Value, interfaces bool) (reflect.Value, error) {
	type pointer struct {
		typ     reflect.Type
		address uintptr
	}
	seen := make(map[pointer]bool)
	for {
		if !value.IsValid() || !value.CanInterface() {
			return reflect.Value{}, fmt.Errorf("receiver is inaccessible")
		}
		if reflectutils.ResolvePointerChain(value.Type()) == nil {
			return reflect.Value{}, fmt.Errorf("cyclic pointer type %v", value.Type())
		}
		switch value.Kind() {
		case reflect.Interface:
			if !interfaces {
				return value, nil
			}
			if value.IsNil() {
				return reflect.Value{}, fmt.Errorf("nil interface of type %v requires an explicit concrete instance", value.Type())
			}
			value = value.Elem()
		case reflect.Pointer:
			var err error
			value, err = allocation.pointer(value)
			if err != nil {
				return reflect.Value{}, err
			}
			key := pointer{typ: value.Type(), address: value.Pointer()}
			if seen[key] {
				return reflect.Value{}, fmt.Errorf("cyclic receiver through %v", value.Type())
			}
			seen[key] = true
			value = value.Elem()
		default:
			return value, nil
		}
	}
}
