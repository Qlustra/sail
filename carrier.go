package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

// carrierPath follows anonymous embeddings to the concrete Sail carrier.
// Breadth-first discovery follows promotion depth without relying on public
// member names, which a composite declaration may shadow.
func carrierPath(typ, carrier reflect.Type) ([]int, error) {
	type embedding struct {
		typ  reflect.Type
		path []int
	}
	typ = utilities.ResolvePointerChain(typ)
	level := []embedding{{typ: typ}}
	seen := map[reflect.Type]int{typ: 0}
	for depth := 0; len(level) > 0; depth++ {
		var found [][]int
		var next []embedding
		for _, entry := range level {
			if entry.typ == carrier {
				found = append(found, entry.path)
				continue
			}
			if entry.typ == nil || entry.typ.Kind() != reflect.Struct {
				continue
			}
			for i := 0; i < entry.typ.NumField(); i++ {
				field := entry.typ.Field(i)
				if !field.Anonymous {
					continue
				}
				embedded := utilities.ResolvePointerChain(field.Type)
				if earlier, visited := seen[embedded]; visited && earlier < depth+1 {
					continue
				}
				seen[embedded] = depth + 1
				next = append(next, embedding{typ: embedded, path: slices.Concat(entry.path, field.Index)})
			}
		}
		if len(found) == 1 {
			return found[0], nil
		}
		if len(found) > 1 {
			return nil, fmt.Errorf("%v: ambiguous embedded carrier %v", typ, carrier)
		}
		level = next
	}
	return nil, fmt.Errorf("%v: missing embedded carrier %v", typ, carrier)
}

// allocateDeclaration only initializes pointers along a selected schema path.
// Subject types are inspected, never allocated here.
func allocateDeclaration(value reflect.Value) (reflect.Value, error) {
	if utilities.ResolvePointerChain(value.Type()) == nil {
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
