package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Bind resolves a populated marker against a live subject and returns its
// receiver as I. The subject must be a nonnil pointer to the recorded root;
// additional pointer wrappers are allowed. I must be an ordinary Go interface;
// Self templates are not specialized during binding.
//
// Binding follows only the recorded field path, using existing storage. It
// never allocates missing subject values, constructs surrogates, or invokes
// subject methods. Addressable concrete fields bind through their address when possible.
// Interface fields expose their current dynamic value, with ordinary Go value
// semantics when that value is not a pointer. Rebind after replacing a pointer
// or interface field to obtain the new receiver.
//
// An existing receiver without I returns present=false with no error. Invalid
// markers or roots, nil path/receiver storage, inaccessible pointer receivers,
// and cyclic receiver resolution return errors. Nil maps and slices are values,
// not missing storage. Method invocation and its errors belong to the caller.
func Bind[I any](subject any, marker decl.MarkerDeclaration) (receiver I, present bool, err error) {
	source, err := markerReference(marker)
	if err != nil {
		return receiver, false, fmt.Errorf("sail: bind: %w", err)
	}
	value, err := bindReference(source, subject)
	if err != nil {
		return receiver, false, fmt.Errorf("%s: %w", source, err)
	}
	receiver, present, err = reflectutils.BindInterface[I](value)
	if err != nil {
		return receiver, false, fmt.Errorf("%s: %w", source, err)
	}
	return receiver, present, nil
}

func markerReference(marker decl.MarkerDeclaration) (decl.Reference, error) {
	value, err := reflectutils.Dereference(reflect.ValueOf(marker))
	if err != nil {
		return decl.Reference{}, fmt.Errorf("marker: %w", err)
	}
	typ := value.Type()
	role, err := recognizeDeclarationRole(typ)
	if err != nil {
		return decl.Reference{}, err
	}
	path, err := reflectutils.EmbeddedPath(typ, role.carrier())
	if err != nil {
		return decl.Reference{}, err
	}
	carrier, err := reflectutils.FieldByPath(value, path)
	if err == nil {
		carrier, err = reflectutils.Dereference(carrier)
	}
	if err != nil {
		return decl.Reference{}, fmt.Errorf("marker %v: %w", typ, err)
	}
	// Read Sail's concrete carrier, bypassing shadowed public accessors and
	// avoiding promoted calls through a nil embedded marker pointer.
	source := decl.ReferenceOf(carrier.Interface())
	if source.Root == nil || source.Type == nil {
		return decl.Reference{}, fmt.Errorf("marker %v has not been populated", typ)
	}
	return source, nil
}

func bindReference(r decl.Reference, subject any) (reflect.Value, error) {
	value := reflect.ValueOf(subject)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return reflect.Value{}, fmt.Errorf("subject must be a nonnil pointer to %v", r.Root)
	}
	if reflectutils.ResolvePointerChain(value.Type()) != reflectutils.ResolvePointerChain(r.Root) {
		return reflect.Value{}, fmt.Errorf("subject root %v does not match recorded root %v", value.Type(), r.Root)
	}
	for value.Type() != r.Root {
		if value.Kind() != reflect.Pointer {
			return reflect.Value{}, fmt.Errorf("subject cannot reach recorded root %v", r.Root)
		}
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("missing root storage: nil %v", value.Type())
		}
		value = value.Elem()
	}
	value, err := reflectutils.FieldByPath(value, r.Path)
	if err != nil {
		return reflect.Value{}, err
	}
	if value.Type() != r.Type {
		return reflect.Value{}, fmt.Errorf("receiving type %v does not match recorded type %v", value.Type(), r.Type)
	}
	return value, nil
}
