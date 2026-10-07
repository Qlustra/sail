package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Populate fills a chart destination from a subject declaration.
func Populate(c *Charter, chart reflect.Value, subjectType reflect.Type) error {
	if c == nil || c.Root == nil {
		return fmt.Errorf("sail: populate requires a compiled charter")
	}

	legendType := reflectutils.ResolvePointerChain(chart.Type())
	if legendType != c.Root.typ {
		return fmt.Errorf("sail: chart type %v does not match charter root %v", chart.Type(), c.Root.typ)
	}

	chart, err := allocateDeclaration(chart)
	if err != nil {
		return err
	}
	p := populator{families: c.families, active: make(map[populationKey]decl.Reference)}
	return p.populate(c.Root, chart, decl.Reference{Root: subjectType, Type: subjectType}, "", nil)
}

type populationKey struct {
	descriptor  *markerDescriptor
	subjectType reflect.Type
}

type populator struct {
	families markFamilies
	active   map[populationKey]decl.Reference
}

func (p *populator) populate(descriptor *markerDescriptor, value reflect.Value, source decl.Reference, structTags reflect.StructTag, qualifiers []string) error {
	if err := descriptor.acceptance.match(source.Type); err != nil {
		return fmt.Errorf("%s as %v (%s): %w", source, descriptor.typ, descriptor.mark, err)
	}
	if err := descriptor.populateMetadata(value, source, structTags, qualifiers); err != nil {
		return fmt.Errorf("%s as %v: %w", source, descriptor.typ, err)
	}
	if descriptor.role != landmarkRole {
		if err := p.populateLanes(descriptor, value, source); err != nil {
			return err
		}
	}
	if err := descriptor.populateQualities(value, source.Type, qualifiers); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	if err := descriptor.populateBarrels(value, source.Type); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	return nil
}

func (p *populator) populateLanes(descriptor *markerDescriptor, value reflect.Value, source decl.Reference) error {
	subjectType := reflectutils.ResolvePointerChain(source.Type)
	if subjectType == nil || subjectType.Kind() != reflect.Struct {
		return fmt.Errorf("%s: %v requires a subject declaration that resolves to a struct, got %v", source, descriptor.typ, source.Type)
	}
	key := populationKey{descriptor: descriptor, subjectType: subjectType}
	if previous, exists := p.active[key]; exists {
		return fmt.Errorf("%s: recursive population of %v against %v returns to %s", source, descriptor.typ, subjectType, previous)
	}
	p.active[key] = source
	defer delete(p.active, key)

	lanes, err := newLanePopulation(descriptor, value, p.families)
	if err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	for i := 0; i < subjectType.NumField(); i++ {
		field := subjectType.Field(i)
		if _, marked := field.Tag.Lookup(descriptor.mark.tag); !marked {
			continue
		}
		child := source.Child(field)
		if !field.IsExported() {
			return fmt.Errorf("%s: marked subject field must be exported", child)
		}
		selected, err := lanes.selectField(field)
		if err != nil {
			return fmt.Errorf("%s: %w", child, err)
		}
		target := selected.candidate.target
		marker := reflect.New(target.typ).Elem()
		if err := p.populate(target, marker, child, field.Tag, selected.qualifiers); err != nil {
			return err
		}
		lanes.store(selected.candidate.spotIndex, marker)
	}
	if err := lanes.finish(); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	return nil
}
