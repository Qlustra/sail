package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	subjectcore "github.com/qlustra/sail/internal/subject"
	"github.com/qlustra/sail/legend"
	"github.com/qlustra/sail/subject"
)

// PopulateMembers allocates a Chart describing the declared member type retained
// by members, using the same declaration interpretation as Populate. It does not
// enumerate members or invoke Enter. An empty typed source is sufficient;
// invalid sources are rejected.
//
// As with Populate, barrel and quality capabilities run on fresh zero-valued
// Subject-type surrogates and must not depend on Subject state.
func PopulateMembers[L legend.Declaration](charter *Charter[L], members subject.Children) (*L, error) {
	memberType, err := subjectcore.MemberType(members)
	if err != nil {
		return nil, fmt.Errorf("sail: populate members: %w", err)
	}

	chart := new(L)
	var compiled *engine.Charter
	if charter != nil {
		compiled = charter.compiled
	}
	if err := engine.Populate(compiled, reflect.ValueOf(chart).Elem(), memberType); err != nil {
		return nil, err
	}
	return chart, nil
}
