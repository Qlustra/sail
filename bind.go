package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// Bind resolves a populated marker against a live subject and returns its
// receiver as I. The subject must be a nonnil pointer to the recorded root;
// additional pointer wrappers are allowed. I must be an ordinary Go interface;
// Self templates are not specialized during binding.
//
// Binding follows only the recorded field path, using existing storage. It
// never allocates missing subjects, constructs surrogates, or invokes subject
// methods. Addressable concrete fields bind through their address when possible.
// Interface fields expose their current dynamic value, with ordinary Go value
// semantics when that value is not a pointer. Rebind after replacing a pointer
// or interface field to obtain the new receiver.
//
// An existing receiver without I returns present=false with no error. Invalid
// markers or roots, nil path/receiver storage, inaccessible pointer receivers,
// and cyclic receiver resolution return errors. Nil maps and slices are values,
// not missing storage. Method invocation and its errors belong to the caller.
func Bind[I any](subject any, occurrence MarkerDeclaration) (receiver I, present bool, err error) {
	source, err := occurrenceReference(occurrence)
	if err != nil {
		return receiver, false, fmt.Errorf("sail: bind: %w", err)
	}
	value, err := source.bind(subject)
	if err != nil {
		return receiver, false, fmt.Errorf("%s: %w", source, err)
	}
	receiver, present, err = utilities.BindInterface[I](value)
	if err != nil {
		return receiver, false, fmt.Errorf("%s: %w", source, err)
	}
	return receiver, present, nil
}

func occurrenceReference(occurrence MarkerDeclaration) (subjectReference, error) {
	value, err := utilities.Dereference(reflect.ValueOf(occurrence))
	if err != nil {
		return subjectReference{}, fmt.Errorf("marker: %w", err)
	}
	typ := value.Type()
	role, err := recognizeDeclarationRole(typ)
	if err != nil {
		return subjectReference{}, err
	}
	path, err := carrierPath(typ, role.carrier())
	if err != nil {
		return subjectReference{}, err
	}
	carrier, err := utilities.FieldByPath(value, path)
	if err == nil {
		carrier, err = utilities.Dereference(carrier)
	}
	if err != nil {
		return subjectReference{}, fmt.Errorf("marker %v: %w", typ, err)
	}
	// Read Sail's concrete carrier, bypassing shadowed public accessors and
	// avoiding promoted calls through a nil embedded marker pointer.
	source := carrier.Interface().(interface{ subjectReference() subjectReference }).subjectReference()
	if source.root == nil || source.typ == nil {
		return subjectReference{}, fmt.Errorf("marker %v has not been populated", typ)
	}
	return source, nil
}

func (r subjectReference) bind(subject any) (reflect.Value, error) {
	value := reflect.ValueOf(subject)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return reflect.Value{}, fmt.Errorf("subject must be a nonnil pointer to %v", r.root)
	}
	if utilities.ResolvePointerChain(value.Type()) != utilities.ResolvePointerChain(r.root) {
		return reflect.Value{}, fmt.Errorf("subject root %v does not match recorded root %v", value.Type(), r.root)
	}
	for value.Type() != r.root {
		if value.Kind() != reflect.Pointer {
			return reflect.Value{}, fmt.Errorf("subject cannot reach recorded root %v", r.root)
		}
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("missing root storage: nil %v", value.Type())
		}
		value = value.Elem()
	}
	value, err := utilities.FieldByPath(value, r.path)
	if err != nil {
		return reflect.Value{}, err
	}
	if value.Type() != r.typ {
		return reflect.Value{}, fmt.Errorf("receiving type %v does not match recorded type %v", value.Type(), r.typ)
	}
	return value, nil
}
