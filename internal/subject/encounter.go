package subject

import (
	"reflect"
	"slices"

	"github.com/qlustra/sail/legend"
)

// Encounter describes a live Subject occurrence through its populated Chart
// marker. Its parent chain includes entered member boundaries. The context
// may be retained while the Chart topology remains stable.
type Encounter struct {
	frame *frame
	facts arrival
}

// Marker returns the original populated marker pointer, or nil for a zero encounter.
func (e Encounter) Marker() legend.MarkerDeclaration {
	if e.frame == nil {
		return nil
	}
	return e.frame.occurrence.Value.Addr().Interface().(legend.MarkerDeclaration)
}

// Parent returns the containing marker encounter. The Subject root has no parent.
func (e Encounter) Parent() (Encounter, bool) {
	if e.frame == nil || e.frame.parent == nil {
		return Encounter{}, false
	}
	return e.frame.parent.encounter(), true
}

// Identity returns this arrival's receiver identity. Its zero value represents
// a zero encounter; compare identities only within their propagation invocation.
func (e Encounter) Identity() Identity { return e.facts.identity }

// First reports the first arrival of this receiver during propagation. It does
// not report whether an interceptor ran or succeeded. Rebinding the same
// receiver after Pass preserves this arrival's original First result.
func (e Encounter) First() bool { return e.facts.first }

// Cycle reports the matching ancestor when this arrival closes a cycle.
// An ancestor scope retains the receiver bound when its descent began, even if
// a later callback replaces that field. Interceptors may inspect, prune, or stop
// the arrival; descent into a cycle is rejected.
func (e Encounter) Cycle() (Encounter, bool) {
	if e.facts.cycle == nil {
		return Encounter{}, false
	}
	return *e.facts.cycle, true
}

// Depth counts marker edges from the Subject root, which is visited at depth zero.
func (e Encounter) Depth() int {
	if e.frame == nil {
		return 0
	}
	return e.frame.depth
}

// Source returns the recorded Subject type and field path, or an empty string
// for a zero encounter.
func (e Encounter) Source() string {
	if e.frame == nil {
		return ""
	}
	return e.frame.prefix + e.frame.occurrence.Source.String()
}

// CarrierType returns the original declared carrier type, including pointers.
// It is nil for a zero encounter.
func (e Encounter) CarrierType() reflect.Type {
	if e.frame == nil {
		return nil
	}
	return e.frame.occurrence.Source.Type
}

// CarrierPath returns a copied field-index path from this Subject or entered
// member root. Source and Parent retain the containing member boundary.
func (e Encounter) CarrierPath() []int {
	if e.frame == nil {
		return nil
	}
	return slices.Clone(e.frame.occurrence.Source.Path)
}

// LanePath returns a copied path to the incoming lane, relative to the parent
// marker. It is nil on the Subject root.
func (e Encounter) LanePath() []int {
	if e.frame == nil {
		return nil
	}
	return slices.Clone(e.frame.occurrence.LanePath)
}

// SpotPath returns a copied path to the incoming spot, relative to its lane.
// It is nil on the Subject root.
func (e Encounter) SpotPath() []int {
	if e.frame == nil {
		return nil
	}
	return slices.Clone(e.frame.occurrence.SpotPath)
}

// ItemIndex identifies the marker within its spot: zero for single and present
// optional spots, the slice index for multi spots, or -1 on the root.
func (e Encounter) ItemIndex() int {
	if e.frame == nil {
		return -1
	}
	return e.frame.occurrence.ItemIndex
}
