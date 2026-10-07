package reflectutils

import (
	"fmt"
	"reflect"
)

// IdentitySignal describes a reliable receiver address, or a value that needs
// containing-field or yielded-occurrence provenance from its traversal frame.
// Keep retains the actual receiver references for as long as the signal is used.
type IdentitySignal struct {
	Type    reflect.Type
	Address uintptr
	Value   bool
	Keep    any
}

// ReceiverIdentity follows the current receiver without creating addressable
// copies. Pointer receivers and their addressable referents share an identity.
// Zero-sized pointer targets are rejected because Go permits distinct targets
// to have equal addresses. Inline zero-sized and unaddressable values instead
// require occurrence provenance supplied by the caller.
//
// Maps and slices use their receiver storage, when available, rather than their
// contents or backing store. Shared backing storage does not identify a receiver.
func ReceiverIdentity(value reflect.Value) (IdentitySignal, error) {
	type pointer struct {
		typ     reflect.Type
		address uintptr
	}
	seen := make(map[pointer]bool)
	var keep []any
	throughPointer := false
	for {
		if !value.IsValid() || !value.CanInterface() {
			return IdentitySignal{}, fmt.Errorf("receiver identity is inaccessible")
		}
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return IdentitySignal{}, fmt.Errorf("receiver identity contains a nil interface of type %v", value.Type())
			}
			keep = append(keep, value.Interface())
			value = value.Elem()
		case reflect.Pointer:
			if value.IsNil() {
				return IdentitySignal{}, fmt.Errorf("receiver identity contains a nil pointer of type %v", value.Type())
			}
			key := pointer{typ: value.Type(), address: value.Pointer()}
			if seen[key] {
				return IdentitySignal{}, fmt.Errorf("cyclic receiver identity through %v", value.Type())
			}
			seen[key] = true
			throughPointer = true
			// Retain snapshots of each pointer, rather than just the original
			// field: handlers may replace that field during propagation.
			keep = append(keep, value.Interface())
			value = value.Elem()
		default:
			if value.Type().Size() == 0 && throughPointer {
				return IdentitySignal{}, fmt.Errorf("receiver identity for pointer target %v is ambiguous because it has zero size", value.Type())
			}
			signal := IdentitySignal{Type: value.Type(), Value: true}
			if value.Type().Size() != 0 && value.CanAddr() && value.Addr().CanInterface() {
				address := value.Addr()
				signal.Address = address.Pointer()
				signal.Value = false
				keep = append(keep, address.Interface())
			} else {
				keep = append(keep, value.Interface())
			}
			signal.Keep = keep
			return signal, nil
		}
	}
}
