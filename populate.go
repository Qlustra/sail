package sail

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// ErrPopulationNotImplemented marks the unfinished application of schema lanes
// to subject fields. Population must not return an empty chart as a success.
var ErrPopulationNotImplemented = errors.New("sail: chart population is not implemented")

// Populate allocates a chart to describe S's static declaration using an
// explored world. It does not construct or modify a subject instance.
func Populate[S any, C chart](world *World[C]) (*C, error) {
	c := new(C)

	err := world.populate(
		reflect.ValueOf(c).Elem(),
		reflect.TypeFor[S](),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (w *World[C]) populate(chart reflect.Value, subject reflect.Type) error {
	if w == nil || w.Root == nil {
		return fmt.Errorf("sail: populate requires an explored world")
	}

	// The subject supplies a declaration, including when S is a pointer type.
	subjectType := utilities.ResolvePointerChain(subject)
	if subjectType == nil || subjectType.Kind() != reflect.Struct {
		return fmt.Errorf("sail: populate subject %v must resolve to a struct", subject)
	}

	chartType := utilities.ResolvePointerChain(chart.Type())
	if chartType != w.Root.typ {
		return fmt.Errorf("sail: populate chart %v does not match world root %v", chart.Type(), w.Root.typ)
	}

	// new(C) owns the destination. If C itself is a pointer, allocate its
	// pointee too, so population starts with the settable chart struct.
	for chart.Kind() == reflect.Pointer {
		if chart.IsNil() {
			chart.Set(reflect.New(chart.Type().Elem()))
		}
		chart = chart.Elem()
	}

	// Application starts at w.Root against subjectType, writing into chart.
	// TODO: follow subject fields in declaration order, select an eligible lane,
	// survey its candidates, and populate each matched spot's marker occurrence
	// before descending. Lane advancement and occurrence storage come next.
	return fmt.Errorf("%v against %v: %w", chart.Type(), subjectType, ErrPopulationNotImplemented)
}
