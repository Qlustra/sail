package subject

import (
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Encounter describes a live Subject occurrence through its populated Chart
// marker. Its parent chain follows declared marker relationships. The context
// may be retained while the Chart topology remains stable.
type Encounter struct {
	occurrence *engine.Occurrence
	depth      int
}

// Marker returns the original populated marker pointer, or nil for a zero encounter.
func (e Encounter) Marker() legend.MarkerDeclaration {
	if e.occurrence == nil {
		return nil
	}
	return e.occurrence.Value.Addr().Interface().(legend.MarkerDeclaration)
}

// Parent returns the containing marker encounter. The Subject root has no parent.
func (e Encounter) Parent() (Encounter, bool) {
	if e.occurrence == nil || e.occurrence.Parent == nil {
		return Encounter{}, false
	}
	return Encounter{occurrence: e.occurrence.Parent, depth: e.depth - 1}, true
}

// Depth counts marker edges from the Subject root, which is visited at depth zero.
func (e Encounter) Depth() int { return e.depth }

// Source returns the recorded Subject type and field path, or an empty string
// for a zero encounter.
func (e Encounter) Source() string {
	if e.occurrence == nil {
		return ""
	}
	return e.occurrence.Source.String()
}

// CarrierType returns the original declared carrier type, including pointers.
// It is nil for a zero encounter.
func (e Encounter) CarrierType() reflect.Type {
	if e.occurrence == nil {
		return nil
	}
	return e.occurrence.Source.Type
}

// CarrierPath returns a copied field-index path from the Subject root.
func (e Encounter) CarrierPath() []int {
	if e.occurrence == nil {
		return nil
	}
	return slices.Clone(e.occurrence.Source.Path)
}

// LanePath returns a copied path to the incoming lane, relative to the parent
// marker. It is nil on the Subject root.
func (e Encounter) LanePath() []int {
	if e.occurrence == nil {
		return nil
	}
	return slices.Clone(e.occurrence.LanePath)
}

// SpotPath returns a copied path to the incoming spot, relative to its lane.
// It is nil on the Subject root.
func (e Encounter) SpotPath() []int {
	if e.occurrence == nil {
		return nil
	}
	return slices.Clone(e.occurrence.SpotPath)
}

// ItemIndex identifies the marker within its spot: zero for single and present
// optional spots, the slice index for multi spots, or -1 on the root.
func (e Encounter) ItemIndex() int {
	if e.occurrence == nil {
		return -1
	}
	return e.occurrence.ItemIndex
}
