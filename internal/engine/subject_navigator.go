package engine

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/reflectutils"
)

// SubjectNavigator follows a populated Chart while resolving its carriers in an
// existing live Subject. Chart topology remains stable; carriers are read again
// when reached so changes to Subject pointers and interfaces remain visible.
type SubjectNavigator struct {
	root        *Occurrence
	carrierRoot func() (reflect.Value, bool, error)
}

// NewSubjectNavigator validates the Chart association and the live Subject root
// before any carrier can be visited. It allocates no Subject values or invokes
// any Subject methods.
func NewSubjectNavigator(charter *Charter, chart reflect.Value, subject any) (*SubjectNavigator, error) {
	root, err := navigationRoot(charter, chart)
	if err != nil {
		return nil, err
	}
	value, err := bindReference(root.Source, subject)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", root.Source, err)
	}
	_, present, err := existingSubjectReceiver(value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", root.Source, err)
	}
	if !present {
		return nil, fmt.Errorf("%s: subject root contains a nil pointer or interface", root.Source)
	}
	return &SubjectNavigator{
		root: root,
		carrierRoot: func() (reflect.Value, bool, error) {
			return subjectCarrierRoot(subject, root.Source.Root)
		},
	}, nil
}

// NewMemberNavigator admits an entered member using its declared value and type.
// Unlike a supplied Subject root, a member may be a value or nil pointer. Values
// keep their original addressability; entering never manufactures a receiver.
func NewMemberNavigator(charter *Charter, chart reflect.Value, member reflect.Value) (*SubjectNavigator, error) {
	root, err := navigationRoot(charter, chart)
	if err != nil {
		return nil, err
	}
	if !member.IsValid() || !member.CanInterface() {
		return nil, fmt.Errorf("%s: entered member is inaccessible", root.Source)
	}
	if member.Type() != root.Source.Root {
		return nil, fmt.Errorf("%s: entered member type %v does not match recorded root %v", root.Source, member.Type(), root.Source.Root)
	}
	return &SubjectNavigator{
		root: root,
		carrierRoot: func() (reflect.Value, bool, error) {
			return member, true, nil
		},
	}, nil
}

// Root returns the populated Legend occurrence, which also participates in
// Subject propagation.
func (n *SubjectNavigator) Root() *Occurrence { return n.root }

// Children returns the next declared Chart occurrences in declaration order.
// It does not inspect live carriers or allocate absent Subject values.
func (n *SubjectNavigator) Children(parent *Occurrence) ([]*Occurrence, error) {
	return immediateChildren(parent)
}

// Carrier returns an occurrence's original declared field value in the Subject.
// A nil pointer or interface anywhere along its path or receiver chain makes the
// branch absent. Nil maps and slices remain present. The value is resolved from
// the Subject root on every call and is never copied to manufacture a receiver.
func (n *SubjectNavigator) Carrier(at *Occurrence) (reflect.Value, bool, error) {
	chain, present, err := n.CarrierChain(at)
	if err != nil || !present {
		return reflect.Value{}, false, err
	}
	return chain[len(chain)-1], true, nil
}

// CarrierChain retains the original root and each declared field value along
// an occurrence's carrier path. Intermediate fields are included even when
// unmarked, so inline identity can be anchored to its actual containing receiver.
// Values are resolved again on each call, with the same presence rules as Carrier.
func (n *SubjectNavigator) CarrierChain(at *Occurrence) ([]reflect.Value, bool, error) {
	if at == nil || at.Source.Root == nil || at.Source.Type == nil {
		return nil, false, fmt.Errorf("carrier requires a populated occurrence")
	}
	if at.Source.Root != n.root.Source.Root {
		return nil, false, fmt.Errorf("carrier root %v does not match navigator root %v", at.Source.Root, n.root.Source.Root)
	}
	value, present, err := n.carrierRoot()
	if err != nil || !present {
		return nil, false, err
	}
	chain := make([]reflect.Value, 1, len(at.Source.Path)+1)
	chain[0] = value
	for _, index := range at.Source.Path {
		value, present, err = existingSubjectPointers(value)
		if err != nil || !present {
			return nil, false, err
		}
		if value.Kind() == reflect.Interface && value.IsNil() {
			return nil, false, nil
		}
		if value.Kind() != reflect.Struct || index < 0 || index >= value.NumField() {
			return nil, false, fmt.Errorf("field index %d cannot be resolved in %v", index, value.Type())
		}
		field := value.Type().Field(index)
		if !field.IsExported() {
			return nil, false, fmt.Errorf("field %v.%s is not exported", value.Type(), field.Name)
		}
		value = value.Field(index)
		chain = append(chain, value)
	}
	if !value.IsValid() || !value.CanInterface() {
		return nil, false, fmt.Errorf("carrier is inaccessible")
	}
	if value.Type() != at.Source.Type {
		return nil, false, fmt.Errorf("carrier type %v does not match recorded type %v", value.Type(), at.Source.Type)
	}
	_, present, err = existingSubjectReceiver(value)
	if err != nil || !present {
		return nil, false, err
	}
	return chain, true, nil
}

func subjectCarrierRoot(subject any, root reflect.Type) (reflect.Value, bool, error) {
	value := reflect.ValueOf(subject)
	if !value.IsValid() || value.Kind() != reflect.Pointer {
		return reflect.Value{}, false, fmt.Errorf("subject must be a nonnil pointer to %v", root)
	}
	terminal := reflectutils.ResolvePointerChain(value.Type())
	expected := reflectutils.ResolvePointerChain(root)
	if terminal == nil || expected == nil {
		return reflect.Value{}, false, fmt.Errorf("subject root has a cyclic pointer type")
	}
	if terminal != expected {
		return reflect.Value{}, false, fmt.Errorf("subject root %v does not match recorded root %v", value.Type(), root)
	}
	for value.Type() != root {
		if value.Kind() != reflect.Pointer {
			return reflect.Value{}, false, fmt.Errorf("subject cannot reach recorded root %v", root)
		}
		if value.IsNil() {
			return reflect.Value{}, false, nil
		}
		value = value.Elem()
	}
	return value, true, nil
}

// existingSubjectPointers preserves interfaces because their dynamic values do
// not introduce declared field paths.
func existingSubjectPointers(value reflect.Value) (reflect.Value, bool, error) {
	if !value.IsValid() || !value.CanInterface() {
		return reflect.Value{}, false, fmt.Errorf("carrier is inaccessible")
	}
	if reflectutils.ResolvePointerChain(value.Type()) == nil {
		return reflect.Value{}, false, fmt.Errorf("cyclic pointer type %v", value.Type())
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false, nil
		}
		value = value.Elem()
	}
	return value, true, nil
}

func existingSubjectReceiver(value reflect.Value) (reflect.Value, bool, error) {
	type pointer struct {
		typ     reflect.Type
		address uintptr
	}
	seen := make(map[pointer]bool)
	for {
		if !value.IsValid() || !value.CanInterface() {
			return reflect.Value{}, false, fmt.Errorf("receiver is inaccessible")
		}
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return reflect.Value{}, false, nil
			}
		case reflect.Pointer:
			if value.IsNil() {
				return reflect.Value{}, false, nil
			}
			key := pointer{typ: value.Type(), address: value.Pointer()}
			if seen[key] {
				return reflect.Value{}, false, fmt.Errorf("cyclic receiver through %v", value.Type())
			}
			seen[key] = true
		default:
			return value, true, nil
		}
		value = value.Elem()
	}
}
