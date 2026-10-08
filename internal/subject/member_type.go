package subject

import "reflect"

// MemberType obtains the retained declaration type without starting iteration.
// It remains internal so Children does not expose reflection in its public API.
func MemberType(children Children) (reflect.Type, error) {
	if err := children.validate(); err != nil {
		return nil, err
	}
	return children.memberType, nil
}
