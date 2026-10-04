package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Populate fills a world destination from an application declaration.
func Populate(c *Chart, world reflect.Value, appType reflect.Type) error {
	if c == nil || c.Root == nil {
		return fmt.Errorf("sail: populate requires an explored chart")
	}

	legendType := reflectutils.ResolvePointerChain(world.Type())
	if legendType != c.Root.typ {
		return fmt.Errorf("sail: world type %v does not match chart root %v", world.Type(), c.Root.typ)
	}

	world, err := allocateDeclaration(world)
	if err != nil {
		return err
	}
	p := populator{families: c.families, active: make(map[populationKey]decl.Reference)}
	return p.populate(c.Root, world, decl.Reference{Root: appType, Type: appType}, "", nil)
}

type populationKey struct {
	descriptor *markerDescriptor
	appType    reflect.Type
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
	appType := reflectutils.ResolvePointerChain(source.Type)
	if appType == nil || appType.Kind() != reflect.Struct {
		return fmt.Errorf("%s: %v requires an application declaration that resolves to a struct, got %v", source, descriptor.typ, source.Type)
	}
	key := populationKey{descriptor: descriptor, appType: appType}
	if previous, exists := p.active[key]; exists {
		return fmt.Errorf("%s: recursive population of %v against %v returns to %s", source, descriptor.typ, appType, previous)
	}
	p.active[key] = source
	defer delete(p.active, key)

	lanes, err := newLanePopulation(descriptor, value, p.families)
	if err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	for i := 0; i < appType.NumField(); i++ {
		field := appType.Field(i)
		if _, marked := field.Tag.Lookup(descriptor.mark.tag); !marked {
			continue
		}
		child := source.Child(field)
		if !field.IsExported() {
			return fmt.Errorf("%s: marked application field must be exported", child)
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
