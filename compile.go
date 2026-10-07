package sail

import (
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// CompileLegend builds a Charter from legend L, without a Subject declaration or instance.
func CompileLegend[L legend.Declaration]() (*Charter[L], error) {
	charter := &Charter[L]{}
	err := charter.Compile()
	return charter, err
}

// Compile publishes a complete Charter only after its declarations have been
// validated. Failed compilation leaves an existing Charter unchanged.
func (c *Charter[L]) Compile() error {
	compiled, err := engine.Explore(reflect.TypeFor[L]())
	if err != nil {
		return err
	}
	c.compiled = compiled
	return nil
}
