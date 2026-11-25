package sail

import (
	"errors"
	"fmt"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/populating"
)

// Navigator holds Sail's reusable compiled grammar, not an instance of the Framework's
// Legend model L. Obtain one through Compile; population fills a separate L.
type Navigator[L any] struct {
	grammar *declarations.Legend
}

// Warning is a non-fatal grammar compilation advisory.
type Warning = declarations.Warning

// Warnings returns a snapshot of compilation advisories. Modifying the returned
// slice does not change the Navigator. A nil or zero Navigator has no warnings.
func (c *Navigator[L]) Warnings() []Warning {
	if c == nil || c.grammar == nil || len(c.grammar.Warnings) == 0 {
		return nil
	}
	return append([]Warning(nil), c.grammar.Warnings...)
}

// Populate uses only the dynamic type of chart, not its field values. It has
// the same struct-root requirement and fresh-result semantics as Build.
// A nil interface has no type evidence; a typed pointer is not a struct root.
func (c *Navigator[L]) Populate(chart any) (*L, error) {
	chartType, check := capture.TypeOf(chart)
	if !check.OK() {
		return nil, fmt.Errorf("cannot capture Chart type: %s", check.Why())
	}
	return c.populate(chartType)
}

func (c *Navigator[L]) newPopulator() *populating.Populator[L] {
	return &populating.Populator[L]{Grammar: c.grammar}
}

// Build interprets the explicit Chart type A and returns a fresh populated
// Legend. A must be a struct type, not a pointer. Build does not construct a
// Chart or return a partially populated Legend on error. A successful Compile
// does not guarantee that a particular Chart is valid.
func (c *Navigator[L]) Build[A any]() (*L, error) {
	return c.populate(capture.Type[A]())
}

func (c *Navigator[L]) populate(chartType *capture.TypeInterrogator) (*L, error) {
	if c == nil || c.grammar == nil {
		return nil, errors.New("cannot populate without a compiled grammar; use Compile")
	}
	return c.newPopulator().Populate(chartType)
}
