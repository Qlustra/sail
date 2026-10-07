package reflectutils

import (
	"fmt"
	"reflect"
	"slices"
)

// EmbeddedPath follows anonymous embeddings to the requested embedded type.
// Breadth-first discovery follows promotion depth without relying on public
// member names, which a composite declaration may shadow.
func EmbeddedPath(typ, target reflect.Type) ([]int, error) {
	type embedding struct {
		typ  reflect.Type
		path []int
	}
	typ = ResolvePointerChain(typ)
	level := []embedding{{typ: typ}}
	seen := map[reflect.Type]int{typ: 0}
	for depth := 0; len(level) > 0; depth++ {
		var found [][]int
		var next []embedding
		for _, entry := range level {
			if entry.typ == target {
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
				embedded := ResolvePointerChain(field.Type)
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
			return nil, fmt.Errorf("%v: ambiguous embedded type %v", typ, target)
		}
		level = next
	}
	return nil, fmt.Errorf("%v: missing embedded type %v", typ, target)
}
