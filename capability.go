package sail

import (
	"reflect"

	"github.com/qlustra/sail/internal/capability"
)

// ReadCapability invokes a declaration method on a fresh zero-valued surrogate
// of carrierType. I must be an interface whose selected method is
// exported, takes no arguments, and returns one value assignable to V.
// Self in I matches the carrier type after pointer normalization, following
// the same signature rules as Quality.
//
// An unavailable or incompatible capability returns present=false with no
// error. Nil types, cyclic pointer chains, and interface declarations cannot
// supply a surrogate and likewise return present=false. Invalid capability
// requests, incompatible result types, and hook panics return errors.
// A successful zero or nil result still returns present=true. Use V=any to
// retain the dynamic type of a Self result.
//
// No live subject is inspected or mutated. The method must not depend on
// initialized subject state. Each call invokes independently; results are
// not cached. Qualities reuse this machinery and require capability presence
// when their owning qualifier is active.
func ReadCapability[V, I any](carrierType reflect.Type, method string) (value V, present bool, err error) {
	reader, err := capability.CompileReader(reflect.TypeFor[V](), reflect.TypeFor[I](), method)
	if err != nil {
		return value, false, err
	}
	resolvedType, unavailable := reader.Resolve(carrierType)
	if unavailable != nil {
		return value, false, nil
	}
	result, err := reader.Read(resolvedType)
	if err != nil {
		return value, false, err
	}
	// Preserve nil interface results and the compiled assignability contract.
	reflect.ValueOf(&value).Elem().Set(result)
	return value, true, nil
}
