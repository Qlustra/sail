package sail

import (
	"reflect"

	"github.com/qlustra/sail/chart"
	chartcore "github.com/qlustra/sail/internal/chart"
	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Charter holds the reusable descriptors compiled from legend L.
// Populate uses it to interpret a Subject declaration into a fresh Chart.
// Compiled descriptors remain internal; callers use the Charter through Sail operations.
type Charter[L legend.Declaration] struct {
	compiled *engine.Charter
}

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

// Traverse resolves the entire Course against the populated Chart before invoking
// any visitor.
// Every waypoint must select exactly one immediate child marker. The Legend root
// anchors navigation but is not itself visited. Selected occurrences are visited
// in route order, ending at the destination. Each encounter passes through every
// registered visitor in list order. Visitor registrations are copied on entry.
//
// Selectors must be read-only and Chart topology must remain stable throughout
// planning and execution. Sail does not detect mutations. Callback panics
// propagate; returned visitor errors retain their cause and encounter context.
// This operation neither traverses a live Subject nor invokes its hooks.
func (c *Charter[L]) Traverse(view *L, traversal chart.Traversal) error {
	var compiled *engine.Charter
	if c != nil {
		compiled = c.compiled
	}
	return chartcore.Traverse(compiled, reflect.ValueOf(view), traversal)
}
