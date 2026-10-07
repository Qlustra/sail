package engine

import (
	"fmt"
	"reflect"
)

// MemberChart interprets a declared member type using the current Charter's
// Legend. Its result can be reused for members of the same declared type,
// including a collection with no live members.
func MemberChart(charter *Charter, memberType reflect.Type) (reflect.Value, error) {
	if charter == nil || charter.Root == nil {
		return reflect.Value{}, fmt.Errorf("entered members require a compiled charter")
	}
	if memberType == nil {
		return reflect.Value{}, fmt.Errorf("entered members require a declared member type")
	}
	chart := reflect.New(charter.Root.typ).Elem()
	if err := Populate(charter, chart, memberType); err != nil {
		return reflect.Value{}, err
	}
	return chart, nil
}
