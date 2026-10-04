package sail

import (
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Explore builds a chart with legend L, without an application declaration or instance.
func Explore[L legend.Declaration]() (*Chart[L], error) {
	chart := &Chart[L]{}
	err := chart.Explore()
	return chart, err
}

// Explore publishes a complete chart only after its declarations have been
// validated. Failed exploration leaves an existing chart unchanged.
func (c *Chart[L]) Explore() error {
	compiled, err := engine.Explore(reflect.TypeFor[L]())
	if err != nil {
		return err
	}
	c.compiled = compiled
	return nil
}
