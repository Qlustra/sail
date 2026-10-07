package subject

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/engine"
)

// Propagation offers live Subject occurrences to ordered interceptors. An empty
// list descends through every present occurrence without invoking callbacks.
type Propagation struct {
	Interceptors Interceptors
}

// Propagate visits the Subject root and its declared occurrences in preorder,
// preserving declaration order. Registrations and the live root are validated
// before callbacks. Interceptor registrations are copied on entry.
//
// Carriers are resolved when reached, including after Pass. Parent callbacks
// may initialize or replace children before Descend. Nil pointer and interface
// branches are skipped; nil maps and slices remain valid receivers. Sail does
// not allocate Subject values or invoke hooks beyond the supplied callbacks.
//
// Chart topology must remain stable. Callback panics propagate; returned errors
// retain their cause with occurrence and interceptor context.
func Propagate(charter *engine.Charter, chart reflect.Value, subject any, propagation Propagation) error {
	interceptors := slices.Clone(propagation.Interceptors)
	if err := interceptors.validate(); err != nil {
		return fmt.Errorf("sail: propagate: %w", err)
	}
	navigator, err := engine.NewSubjectNavigator(charter, chart, subject)
	if err != nil {
		return fmt.Errorf("sail: propagate: %w", err)
	}
	p := propagator{navigator: navigator, interceptors: interceptors}
	_, err = p.visit(Encounter{occurrence: navigator.Root()})
	return err
}

type propagator struct {
	navigator    *engine.SubjectNavigator
	interceptors Interceptors
}

func (p *propagator) visit(at Encounter) (stopped bool, err error) {
	flow, err := p.interceptors.intercept(p.navigator, at)
	if err != nil {
		return false, fmt.Errorf("sail: propagate at %s: %w", at.Source(), err)
	}
	switch flow {
	case Stop:
		return true, nil
	case Prune:
		return false, nil
	}
	children, err := p.navigator.Children(at.occurrence)
	if err != nil {
		return false, fmt.Errorf("sail: propagate at %s: %w", at.Source(), err)
	}
	for _, child := range children {
		stopped, err := p.visit(Encounter{occurrence: child, depth: at.depth + 1})
		if err != nil || stopped {
			return stopped, err
		}
	}
	return false, nil
}
