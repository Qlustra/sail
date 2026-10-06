package sail

import (
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// CompileLegend builds a chart with legend L, without an application declaration or instance.
func CompileLegend[L legend.Declaration]() (*Chart[L], error) {
	chart := &Chart[L]{}
	err := chart.Compile()
	return chart, err
}

// Compile publishes a complete chart only after its declarations have been
// validated. Failed compilation leaves an existing chart unchanged.
func (c *Chart[L]) Compile() error {
	compiled, err := engine.Explore(reflect.TypeFor[L]())
	if err != nil {
		return err
	}
	c.compiled = compiled
	return nil
}
