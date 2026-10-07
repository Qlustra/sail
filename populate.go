package sail

import (
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Populate allocates a Chart using legend L's shape to describe S's declaration.
// The Charter supplies the compiled descriptors; no live Subject is required.
// Barrel and quality capabilities run on fresh zero-valued Subject-type
// surrogates and must not depend on Subject state.
func Populate[S any, L legend.Declaration](charter *Charter[L]) (*L, error) {
	chart := new(L)
	var compiled *engine.Charter
	if charter != nil {
		compiled = charter.compiled
	}
	if err := engine.Populate(compiled, reflect.ValueOf(chart).Elem(), reflect.TypeFor[S]()); err != nil {
		return nil, err
	}
	return chart, nil
}
