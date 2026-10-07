package subject

import (
	"fmt"
	"iter"
	"reflect"
)

// Enterable supplies live member occurrences when propagation descends through
// a reached carrier. The carrier owns discovery, composition, keys, and order.
type Enterable interface {
	Enter() (Children, error)
}

// Children is an opaque typed source of live members. Construct one with Iterate.
// Its zero value is invalid.
type Children struct {
	memberType reflect.Type
	iterate    iter.Seq2[any, reflect.Value]
	err        error
}

// Iterate retains the declared member type even when the source is empty.
// Yielded values preserve ordinary Go value semantics: Sail does not take the
// address of a yielded copy to manufacture pointer receivers. Iterator order
// and early termination are preserved. A nil iterator is reported by propagation.
func Iterate[K, T any](members iter.Seq2[K, T]) Children {
	children := Children{memberType: reflect.TypeFor[T]()}
	if members == nil {
		children.err = fmt.Errorf("Iterate requires a nonnil member iterator")
		return children
	}
	children.iterate = func(yield func(any, reflect.Value) bool) {
		members(func(key K, member T) bool {
			return yield(key, reflect.ValueOf(member))
		})
	}
	return children
}

func (children Children) validate() error {
	if children.err != nil {
		return children.err
	}
	if children.memberType == nil || children.iterate == nil {
		return fmt.Errorf("children require a typed member iterator; construct them with Iterate")
	}
	return nil
}
