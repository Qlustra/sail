package sail

import "reflect"

// Self is a capability signature placeholder for the receiving application type.
// Quality, ReadCapability, and AcceptContract resolve application pointers before
// matching Self, retaining the identity of the defined type. Self may appear as a
// method's parameter or result; compound types containing Self are unsupported.
// It is a contract placeholder, never the stored value type.
type Self struct{}

// ReadCapability invokes a declaration method on a fresh zero-valued surrogate
// of receivingType. I must be an interface whose selected method is
// exported, takes no arguments, and returns one value assignable to V.
// Self in I matches the receiving type after pointer normalization, following
// the same signature rules as Quality.
//
// An unavailable or incompatible capability returns present=false with no
// error. Nil types, cyclic pointer chains, and interface declarations cannot
// supply a surrogate and likewise return present=false. Invalid capability
// requests, incompatible result storage, and hook panics return errors.
// A successful zero or nil result still returns present=true. Use V=any to
// retain the dynamic type of a Self result.
//
// No live application is inspected or mutated. The method must not depend on
// initialized application state. Each call invokes independently; results are
// not cached. Qualities reuse this machinery and require capability presence
// when their owning qualifier is active.
func ReadCapability[V, I any](receivingType reflect.Type, method string) (value V, present bool, err error) {
	reader, err := compileCapabilityReader(reflect.TypeFor[V](), reflect.TypeFor[I](), method)
	if err != nil {
		return value, false, err
	}
	receiver, unavailable := reader.resolve(receivingType)
	if unavailable != nil {
		return value, false, nil
	}
	result, err := reader.read(receiver)
	if err != nil {
		return value, false, err
	}
	// Reflection assignment preserves nil interface results and applies the
	// checked assignability contract without a potentially failing assertion.
	reflect.ValueOf(&value).Elem().Set(result)
	return value, true, nil
}
