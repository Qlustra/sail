package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/chart"
	"github.com/qlustra/sail/internal/engine"
	subjectcore "github.com/qlustra/sail/internal/subject"
	"github.com/qlustra/sail/legend"
	"github.com/qlustra/sail/subject"
)

// Vessel associates a compiled Charter and populated Chart with a live Subject.
// The Chart describes the Subject's declaration; Subject holds its actual state.
// Multiple Vessels may share a Charter, Chart, or Subject.
type Vessel[S any, L legend.Declaration] struct {
	charter *Charter[L]
	chart   *L
	Subject *S
}

// Materialize creates a Vessel from a compiled Charter and populated Chart.
// Omitting subject allocates a fresh zero-valued Subject with new(S).
// Supplying one nonnil pointer retains that exact Subject, without copying it.
// More than one pointer, explicit nil, or an incompatible association is an error.
// Materialization does not allocate nested pointers or invoke Subject hooks;
// framework composition and initialization remain the caller's responsibility.
func Materialize[S any, L legend.Declaration](charter *Charter[L], chart *L, subject ...*S) (*Vessel[S, L], error) {
	if len(subject) > 1 {
		return nil, fmt.Errorf("sail: materialize accepts at most one Subject, got %d", len(subject))
	}
	var value *S
	if len(subject) == 0 {
		value = new(S)
	} else {
		value = subject[0]
		if value == nil {
			return nil, fmt.Errorf("sail: materialize requires a nonnil Subject when supplied")
		}
	}
	var compiled *engine.Charter
	if charter != nil {
		compiled = charter.compiled
	}
	if err := engine.ValidateVessel(compiled, reflect.ValueOf(chart), value); err != nil {
		return nil, fmt.Errorf("sail: materialize: %w", err)
	}
	return &Vessel[S, L]{charter: charter, chart: chart, Subject: value}, nil
}

// Traverse follows a Course through the Vessel's Chart. Visitors receive
// populated markers; traversal does not walk or invoke the live Subject.
func (v *Vessel[S, L]) Traverse(traversal chart.Traversal) error {
	if v == nil {
		return fmt.Errorf("sail: traverse requires a Vessel")
	}
	return v.charter.Traverse(v.chart, traversal)
}

// Propagate visits the live Subject using an explicit initial order. Interceptors
// bind existing receivers and choose local descent order, pruning, or stopping.
// Enterable carriers supply typed members interpreted through this Vessel's
// Legend. DescendSelected replaces normal children for one occurrence. Nil
// branches are skipped; Chart topology must remain stable.
//
// Shared receivers are offered on each arrival. First identifies the first
// arrival, and Cycle identifies a return to an ancestor. Cyclic arrivals may be
// handled, but descent into a cycle fails. Errors stop immediately and retain
// their cause and source; Stop, errors, and panics cancel pending completions.
func (v *Vessel[S, L]) Propagate(order subject.Order, propagation subject.Propagation) error {
	if v == nil {
		return fmt.Errorf("sail: propagate requires a Vessel")
	}
	var compiled *engine.Charter
	if v.charter != nil {
		compiled = v.charter.compiled
	}
	return subjectcore.Propagate(compiled, reflect.ValueOf(v.chart), v.Subject, order, propagation)
}
