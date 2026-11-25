// Package values holds occurrence-local model state, not compiled metadata.
package values

import (
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/logic"
)

type Value[T any] struct {
	value   T
	present bool
}

func (v Value[T]) Get() T        { return v.value }
func (v Value[T]) Present() bool { return v.present }

func (v *Value[T]) fill(value any) logic.Check {
	if v == nil {
		return logic.Nope("invalid_destination", "nil value cell")
	}
	result, check := capture.Assign[T](value)
	if !check.OK() {
		return check
	}
	v.value, v.present = result, true
	return logic.Ok()
}

type Presence struct{ present bool }

func (p Presence) Present() bool { return p.present }
func (p *Presence) mark() logic.Check {
	if p == nil {
		return logic.Nope("invalid_destination", "nil presence cell")
	}
	p.present = true
	return logic.Ok()
}

// Fill stores an assignable result in a writable model occurrence.
func Fill(destination any, value any) (check logic.Check) {
	defer func() {
		if recover() != nil {
			check = logic.Nope("invalid_destination", "unwritable destination")
		}
	}()
	if inspect.IsNil(destination) {
		return logic.Nope("invalid_destination", "nil destination")
	}
	w, ok := destination.(interface{ fill(any) logic.Check })
	if !ok {
		return logic.Nope("invalid_destination", "destination is not writable value state")
	}
	return w.fill(value)
}

// Mark records a presence-only model occurrence.
func Mark(destination any) (check logic.Check) {
	defer func() {
		if recover() != nil {
			check = logic.Nope("invalid_destination", "unwritable destination")
		}
	}()
	if inspect.IsNil(destination) {
		return logic.Nope("invalid_destination", "nil destination")
	}
	w, ok := destination.(interface{ mark() logic.Check })
	if !ok {
		return logic.Nope("invalid_destination", "destination is not writable presence state")
	}
	return w.mark()
}
