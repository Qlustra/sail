package subject

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"

	"github.com/qlustra/sail/internal/engine"
)

// Propagation offers each occurrence to interceptors in registration order.
// An empty list descends automatically using the explicitly chosen visit order.
type Propagation []Interceptor

// Propagate follows a stable Chart through live receivers. Carriers and identity
// facts are resolved on arrival, after Pass, and before descent. Interceptors may
// change Subject state. Stop, errors, and panics cancel pending completions;
// panics propagate. Returned errors preserve their cause and occurrence context.
func Propagate(charter *engine.Charter, chart reflect.Value, subject any, order Order, propagation Propagation) error {
	if err := order.validate(); err != nil {
		return fmt.Errorf("sail: propagate: %w", err)
	}
	registrations := slices.Clone(propagation)
	if err := registrations.validate(); err != nil {
		return fmt.Errorf("sail: propagate: %w", err)
	}
	navigator, err := engine.NewSubjectNavigator(charter, chart, subject)
	if err != nil {
		return fmt.Errorf("sail: propagate: %w", err)
	}
	p := propagator{
		charter: charter, interceptors: registrations,
		charts: make(map[reflect.Type]reflect.Value),
		seen:   make(map[Identity]struct{}), origins: make(map[fieldOrigin]*valueOrigin),
		cursors: make(map[*childCursor]struct{}),
	}
	defer func() {
		p.closeCursors()
		runtime.KeepAlive(p.keep)
	}()
	root := newFrame(navigator, navigator.Root(), nil, "", &valueOrigin{})
	return p.walk(root, order)
}

type frame struct {
	navigator  *engine.SubjectNavigator
	occurrence *engine.Occurrence
	parent     *frame
	depth      int
	prefix     string
	origin     *valueOrigin
	arrivals   map[Identity]bool
	current    arrival
	order      Order
	complete   func() error
	pending    int
	closed     bool
	finished   bool
}

func newFrame(navigator *engine.SubjectNavigator, occurrence *engine.Occurrence, parent *frame, prefix string, origin *valueOrigin) *frame {
	f := &frame{navigator: navigator, occurrence: occurrence, parent: parent, prefix: prefix, origin: origin, arrivals: make(map[Identity]bool)}
	if parent != nil {
		f.depth = parent.depth + 1
	}
	return f
}

func (f *frame) encounter() Encounter { return Encounter{frame: f, facts: f.current} }
func (f *frame) failure(err error) error {
	return fmt.Errorf("sail: propagate at %s: %w", f.encounter().Source(), err)
}

type propagator struct {
	charter      *engine.Charter
	interceptors Propagation
	charts       map[reflect.Type]reflect.Value
	seen         map[Identity]struct{}
	origins      map[fieldOrigin]*valueOrigin
	keep         []any
	cursors      map[*childCursor]struct{}
	stopped      bool
}

func (p *propagator) arrive(f *frame, inherited Order) (Route, error) {
	for index, interceptor := range p.interceptors {
		value, present, err := p.carrier(f)
		if err != nil {
			return Route{}, err
		}
		if !present {
			return Prune(), nil
		}
		route, err := interceptor.invoke(value, f.encounter())
		if err != nil {
			return Route{}, fmt.Errorf("interceptor %d: %w", index+1, err)
		}
		if err := route.validate(); err != nil {
			return Route{}, fmt.Errorf("interceptor %d: %w", index+1, err)
		}
		if route.action != passRoute {
			return route, nil
		}
	}
	// Empty registrations and all-Pass descent still establish identity facts.
	_, present, err := p.carrier(f)
	if err != nil {
		return Route{}, err
	}
	if !present {
		return Prune(), nil
	}
	return Descend(inherited), nil
}

func (p *propagator) carrier(f *frame) (reflect.Value, bool, error) {
	value, present, err := f.navigator.Carrier(f.occurrence)
	if err != nil || !present {
		return value, present, err
	}
	_, err = p.identify(f, value)
	return value, true, err
}

// begin returns a child cursor only for an accepted, present descent. Refreshing
// here observes replacements made by the accepting callback before checking cycles.
func (p *propagator) begin(f *frame, inherited Order) (*childCursor, error) {
	route, err := p.arrive(f, inherited)
	if err != nil {
		return nil, f.failure(err)
	}
	switch route.action {
	case stopRoute:
		p.stopped = true
		return nil, nil
	case pruneRoute:
		f.closed = true
		return nil, p.finish(f)
	}
	value, present, err := p.carrier(f)
	if err != nil {
		return nil, f.failure(err)
	}
	f.order, f.complete = route.order, route.complete
	if !present {
		f.closed = true
		return nil, p.finish(f)
	}
	if f.current.cycle != nil {
		return nil, f.failure(fmt.Errorf("cannot descend into a cycle returning to %s", f.current.cycle.Source()))
	}
	cursor, err := p.children(f, value, route.selected)
	if err != nil {
		return nil, f.failure(err)
	}
	return cursor, nil
}

func (p *propagator) walk(f *frame, inherited Order) error {
	cursor, err := p.begin(f, inherited)
	if err != nil || p.stopped || cursor == nil {
		return err
	}
	if f.order == DepthFirst {
		return p.depth(cursor)
	}
	return p.breadth(cursor)
}

func (p *propagator) depth(cursor *childCursor) error {
	defer p.closeCursor(cursor)
	for !p.stopped {
		child, present, err := p.next(cursor)
		if err != nil || !present {
			return err
		}
		if err := p.walk(child, cursor.parent.order); err != nil {
			return err
		}
	}
	return nil
}

// A breadth scope keeps child sources lazy. Different local orders drain their
// own scope before this frontier resumes; same-order sources join the next level.
func (p *propagator) breadth(root *childCursor) error {
	frontier := []*childCursor{root}
	for len(frontier) > 0 && !p.stopped {
		var next []*childCursor
		for _, cursor := range frontier {
			for !p.stopped {
				child, present, err := p.next(cursor)
				if err != nil {
					return err
				}
				if !present {
					break
				}
				children, err := p.begin(child, BreadthFirst)
				if err != nil || p.stopped {
					return err
				}
				if children == nil {
					continue
				}
				if child.order == DepthFirst {
					if err := p.depth(children); err != nil {
						return err
					}
				} else {
					next = append(next, children)
				}
			}
		}
		frontier = next
	}
	return nil
}

// finish closes successful scopes from the inside out. A breadth-first parent
// waits for source exhaustion as well as all child scopes, including overrides.
func (p *propagator) finish(f *frame) error {
	if p.stopped || f.finished || !f.closed || f.pending != 0 {
		return nil
	}
	f.finished = true
	if f.complete != nil {
		if err := f.complete(); err != nil {
			return f.failure(fmt.Errorf("completion: %w", err))
		}
	}
	if f.parent != nil {
		f.parent.pending--
		return p.finish(f.parent)
	}
	return nil
}
