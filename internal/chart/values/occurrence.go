package values

import (
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/logic"
)

// ChartLocation identifies a Chart field relative to its exact root type.
type ChartLocation struct {
	chartType *capture.TypeInterrogator
	index     []int
}

// ChartType returns an independent interrogator for the recorded root type.
func (l ChartLocation) ChartType() *capture.TypeInterrogator {
	if l.chartType == nil || l.chartType.TypeInspector == nil {
		return nil
	}
	return capture.TypeReflection(l.chartType.CapturedType())
}

// Index returns a copy of the root-relative field index path.
func (l ChartLocation) Index() []int { return append([]int(nil), l.index...) }

// Resolve captures the actual field in a supplied live Chart. The instance must
// be a non-nil pointer to the exact recorded root type; no value is copied or
// constructed. The returned handle references that live field for its lifetime.
// Neither the location nor its populated Legend retains the supplied instance.
func (l *ChartLocation) Resolve(chart any) (*capture.ValueInterrogator, logic.Check) {
	if l == nil || l.chartType == nil || len(l.index) == 0 {
		return nil, logic.Nope("recipe_location_unavailable", "no Chart location was recorded")
	}
	return capture.ResolveField(chart, l.chartType, l.index)
}

// Occurrence holds correspondence for one populated model occurrence.
type Occurrence struct{ location *ChartLocation }

// Present reports whether correspondence was recorded.
func (o Occurrence) Present() bool { return o.location != nil }

func (o Occurrence) storedLocation() *ChartLocation { return o.location }

// StoredLocation reads the marker cell, not an application method with the
// same name as a public reader. A nil location denotes an absent occurrence.
func StoredLocation(token any) (*ChartLocation, bool) {
	o, ok := token.(interface{ storedLocation() *ChartLocation })
	if !ok {
		return nil, false
	}
	return o.storedLocation(), true
}

// ChartLocation returns a snapshot of the recorded correspondence.
func (o Occurrence) ChartLocation() (*ChartLocation, logic.Check) {
	if o.location == nil {
		return nil, logic.Nope("recipe_location_unavailable", "no Chart location was recorded")
	}
	return &ChartLocation{chartType: o.location.ChartType(), index: o.location.Index()}, logic.Ok()
}

func (o *Occurrence) recordOccurrence(chartType *capture.TypeInterrogator, index []int) logic.Check {
	if o == nil {
		return logic.Nope("invalid_destination", "nil occurrence state")
	}
	if chartType == nil || chartType.TypeInspector == nil || chartType.CapturedType() == nil {
		return logic.Nope("invalid_recipe_location", "missing Chart root type")
	}
	root := capture.TypeReflection(chartType.CapturedType())
	if check := root.IsStructKind(); !check.OK() {
		return check
	}
	if len(index) == 0 {
		return logic.Nope("invalid_recipe_location", "empty Chart field path")
	}
	for _, part := range index {
		if part < 0 {
			return logic.Nope("invalid_recipe_location", "negative Chart field index")
		}
	}
	o.location = &ChartLocation{chartType: root, index: append([]int(nil), index...)}
	return logic.Ok()
}

// RecordOccurrence writes checked correspondence to a writable model occurrence.
func RecordOccurrence(destination any, chartType *capture.TypeInterrogator, index []int) (check logic.Check) {
	defer func() {
		if recover() != nil {
			check = logic.Nope("invalid_destination", "unwritable destination")
		}
	}()
	if inspect.IsNil(destination) {
		return logic.Nope("invalid_destination", "nil destination")
	}
	w, ok := destination.(interface {
		recordOccurrence(*capture.TypeInterrogator, []int) logic.Check
	})
	if !ok {
		return logic.Nope("invalid_destination", "destination is not writable occurrence state")
	}
	return w.recordOccurrence(chartType, index)
}
