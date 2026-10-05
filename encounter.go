package sail

import (
	"slices"

	"github.com/qlustra/sail/internal/engine"
	"github.com/qlustra/sail/legend"
)

// Encounter describes one selected arrival in a World. Its parent chain follows
// marker relationships, with lanes and spots recorded as incoming edge context.
// The context may be retained while the World's topology remains stable.
type Encounter struct {
	occurrence *engine.Occurrence
	step       int
	length     int
}

// Marker returns the original populated marker pointer, or nil for a zero encounter.
func (e Encounter) Marker() legend.MarkerDeclaration {
	if e.occurrence == nil {
		return nil
	}
	return e.occurrence.Value.Addr().Interface().(legend.MarkerDeclaration)
}

// Parent returns the preceding marker encounter. The first selected marker's
// parent is the Legend root, which anchors navigation without being visited.
func (e Encounter) Parent() (Encounter, bool) {
	if e.occurrence == nil || e.occurrence.Parent == nil {
		return Encounter{}, false
	}
	return Encounter{occurrence: e.occurrence.Parent, step: e.step - 1, length: e.length}, true
}

// Depth counts marker edges from the Legend root (depth zero).
func (e Encounter) Depth() int {
	if e.occurrence == nil {
		return 0
	}
	return e.step + 1
}

// Step is the zero-based Course waypoint index, or -1 for the root or a zero encounter.
func (e Encounter) Step() int {
	if e.occurrence == nil {
		return -1
	}
	return e.step
}

// Destination reports whether this is the final selected marker of the Course.
func (e Encounter) Destination() bool {
	return e.occurrence != nil && e.step >= 0 && e.step == e.length-1
}

// LanePath returns a copied field-index path to the incoming lane, relative to
// the parent marker. It is nil on the Legend root.
func (e Encounter) LanePath() []int {
	if e.occurrence == nil {
		return nil
	}
	return slices.Clone(e.occurrence.LanePath)
}

// SpotPath returns a copied field-index path to the incoming spot, relative to
// its lane. It is nil on the Legend root.
func (e Encounter) SpotPath() []int {
	if e.occurrence == nil {
		return nil
	}
	return slices.Clone(e.occurrence.SpotPath)
}

// ItemIndex identifies the marker within its spot's Value: zero for single and
// present optional spots, the slice index for multi spots, or -1 on the root.
func (e Encounter) ItemIndex() int {
	if e.occurrence == nil {
		return -1
	}
	return e.occurrence.ItemIndex
}
