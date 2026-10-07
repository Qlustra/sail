package subject

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/internal/reflectutils"
)

type childCandidate struct {
	occurrence *engine.Occurrence
	key        any
	member     reflect.Value
	entered    bool
}

type childCursor struct {
	parent *frame
	chart  reflect.Value
	next   func() (childCandidate, bool)
	stop   func()
}

func (p *propagator) children(f *frame, value reflect.Value, selected *Children) (*childCursor, error) {
	var declared []*engine.Occurrence
	var members Children
	var chart reflect.Value
	entered := selected != nil
	sourceName := "selected children"
	if selected != nil {
		members = *selected
	} else {
		sourceName = "enter"
		var err error
		declared, err = f.navigator.Children(f.occurrence)
		if err != nil {
			return nil, err
		}
		source, present, err := reflectutils.BindInterface[Enterable](value)
		if err != nil {
			return nil, fmt.Errorf("enter: %w", err)
		}
		entered = present
		if entered {
			members, err = source.Enter()
			if err != nil {
				return nil, fmt.Errorf("enter: %w", err)
			}
		}
	}
	if entered {
		if err := members.validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", sourceName, err)
		}
		chart = p.charts[members.memberType]
		if !chart.IsValid() {
			var err error
			chart, err = engine.MemberChart(p.charter, members.memberType)
			if err != nil {
				return nil, fmt.Errorf("entered member %v: %w", members.memberType, err)
			}
			p.charts[members.memberType] = chart
		}
	}
	sequence := func(yield func(childCandidate) bool) {
		for _, occurrence := range declared {
			if !yield(childCandidate{occurrence: occurrence}) {
				return
			}
		}
		if entered {
			members.iterate(func(key any, member reflect.Value) bool {
				return yield(childCandidate{key: key, member: member, entered: true})
			})
		}
	}
	next, stop := iter.Pull(sequence)
	cursor := &childCursor{parent: f, chart: chart, next: next, stop: stop}
	p.cursors[cursor] = struct{}{}
	return cursor, nil
}

func (p *propagator) next(cursor *childCursor) (*frame, bool, error) {
	candidate, present := cursor.next()
	if !present {
		p.closeCursor(cursor)
		cursor.parent.closed = true
		return nil, false, p.finish(cursor.parent)
	}
	parent := cursor.parent
	var child *frame
	if candidate.entered {
		navigator, err := engine.NewMemberNavigator(p.charter, cursor.chart, candidate.member)
		if err != nil {
			return nil, false, parent.failure(fmt.Errorf("entered member [%v]: %w", candidate.key, err))
		}
		prefix := fmt.Sprintf("%s[%v] -> ", parent.encounter().Source(), candidate.key)
		child = newFrame(navigator, navigator.Root(), parent, prefix, &valueOrigin{})
	} else {
		child = newFrame(parent.navigator, candidate.occurrence, parent, parent.prefix, parent.origin)
	}
	parent.pending++
	return child, true, nil
}

func (p *propagator) closeCursor(cursor *childCursor) {
	if _, open := p.cursors[cursor]; open {
		delete(p.cursors, cursor)
		cursor.stop()
	}
}
func (p *propagator) closeCursors() {
	for cursor := range p.cursors {
		p.closeCursor(cursor)
	}
}
