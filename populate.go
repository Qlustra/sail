package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

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

	chartType := utilities.ResolvePointerChain(chart.Type())
	if chartType != w.Root.typ {
		return fmt.Errorf("sail: populate chart %v does not match world root %v", chart.Type(), w.Root.typ)
	}

	chart, err := allocateDeclaration(chart)
	if err != nil {
		return err
	}
	p := populator{families: w.families, active: make(map[populationKey]subjectReference)}
	return p.populate(w.Root, chart, subjectReference{root: subject, typ: subject}, "", nil)
}

type populationKey struct {
	location *location
	subject  reflect.Type
}

type populator struct {
	families markFamilies
	active   map[populationKey]subjectReference
}

func (p *populator) populate(location *location, value reflect.Value, source subjectReference, tags reflect.StructTag, qualifiers []string) error {
	if err := location.populateMetadata(value, source, tags, qualifiers); err != nil {
		return fmt.Errorf("%s as %v: %w", source, location.typ, err)
	}
	if location.role == landmarkRole {
		return nil
	}
	subject := utilities.ResolvePointerChain(source.typ)
	if subject == nil || subject.Kind() != reflect.Struct {
		return fmt.Errorf("%s: %v requires a subject declaration that resolves to a struct, got %v", source, location.typ, source.typ)
	}
	key := populationKey{location: location, subject: subject}
	if previous, exists := p.active[key]; exists {
		return fmt.Errorf("%s: recursive population of %v against %v returns to %s", source, location.typ, subject, previous)
	}
	p.active[key] = source
	defer delete(p.active, key)

	lanes, err := newLanePopulation(location, value, p.families)
	if err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	for i := 0; i < subject.NumField(); i++ {
		field := subject.Field(i)
		if _, marked := field.Tag.Lookup(location.mark.tag); !marked {
			continue
		}
		child := source.child(field)
		if !field.IsExported() {
			return fmt.Errorf("%s: marked subject field must be exported", child)
		}
		selected, err := lanes.selectField(field)
		if err != nil {
			return fmt.Errorf("%s: %w", child, err)
		}
		target := selected.candidate.target
		occurrence := reflect.New(target.typ).Elem()
		if err := p.populate(target, occurrence, child, field.Tag, selected.qualifiers); err != nil {
			return err
		}
		lanes.store(selected.candidate.spotIndex, occurrence)
	}
	if err := lanes.finish(); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	return nil
}
