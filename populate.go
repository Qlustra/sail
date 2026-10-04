package sail

import (
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Populate allocates a world using legend L's shape to describe A's declaration.
// The chart supplies the explored descriptors; no live application is required.
// Barrel and quality capabilities run on fresh zero-valued application-type
// surrogates and must not depend on application state.
func Populate[A any, L legend.Declaration](chart *Chart[L]) (*L, error) {
	world := new(L)
	var compiled *engine.Chart
	if chart != nil {
		compiled = chart.compiled
	}
	if err := engine.Populate(compiled, reflect.ValueOf(world).Elem(), reflect.TypeFor[A]()); err != nil {
		return nil, err
	}
	return world, nil
}
