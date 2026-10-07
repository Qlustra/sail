package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/engine"
)

// Course is a reusable sequence of immediate-child selections starting at the
// Chart root. It retains neither matched markers nor execution progress.
type Course []Waypoint

// Waypoint selects the next marker in a connected Course. Construct one with
// Select. A zero Waypoint is invalid and is rejected before visiting begins.
type Waypoint struct {
	selection engine.Waypoint
	err       error
}

// Select matches a marker value, pointer, or interface T, optionally narrowed
// by one predicate. Interface matching prefers the marker's pointer.
//
// A predicate receives a shallow copy of the concrete marker (a pointer to that
// copy when T is a pointer or a matching interface). Referenced data remains
// shared under ordinary Go copy semantics. Predicates must be read-only; Sail
// neither deep-copies the Chart nor detects mutations. Invalid predicate lists
// are reported by Traverse rather than panicking here.
func Select[T any](predicates ...func(T) bool) Waypoint {
	waypoint := Waypoint{selection: engine.Waypoint{Type: reflect.TypeFor[T]()}}
	if len(predicates) > 1 {
		waypoint.err = fmt.Errorf("Select[%v] accepts at most one predicate", waypoint.selection.Type)
		return waypoint
	}
	var predicate func(T) bool
	if len(predicates) == 1 {
		predicate = predicates[0]
		if predicate == nil {
			waypoint.err = fmt.Errorf("Select[%v] requires a nonnil predicate when supplied", waypoint.selection.Type)
			return waypoint
		}
	}
	waypoint.selection.Match = func(value reflect.Value) bool {
		if _, matches := markerAs[T](value); !matches {
			return false
		}
		if predicate == nil {
			return true
		}
		copy := reflect.New(value.Type()).Elem()
		copy.Set(value)
		marker, _ := markerAs[T](copy)
		return predicate(marker)
	}
	return waypoint
}

// markerAs adapts an addressable Chart marker without allocating or following
// Subject values. Pointer-first matching keeps interface visitors attached
// to the original marker; concrete value requests use ordinary Go copies.
func markerAs[T any](value reflect.Value) (marker T, matches bool) {
	if marker, matches = value.Addr().Interface().(T); matches {
		return marker, true
	}
	marker, matches = value.Interface().(T)
	return marker, matches
}
