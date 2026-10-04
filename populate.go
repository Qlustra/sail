package sail

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// Populate allocates a world using legend L's shape to describe A's declaration.
// The chart supplies the explored descriptors; no live application is required.
// Barrel and quality capabilities run on fresh zero-valued application-type
// surrogates and must not depend on application state.
func Populate[A any, L legend](chart *Chart[L]) (*L, error) {
	world := new(L)

	err := chart.populate(
		reflect.ValueOf(world).Elem(),
		reflect.TypeFor[A](),
	)
	if err != nil {
		return nil, err
	}

	return world, nil
}

func (c *Chart[L]) populate(world reflect.Value, appType reflect.Type) error {
	if c == nil || c.Root == nil {
		return fmt.Errorf("sail: populate requires an explored chart")
	}

	legendType := utilities.ResolvePointerChain(world.Type())
	if legendType != c.Root.typ {
		return fmt.Errorf("sail: world type %v does not match chart root %v", world.Type(), c.Root.typ)
	}

	world, err := allocateDeclaration(world)
	if err != nil {
		return err
	}
	p := populator{families: c.families, active: make(map[populationKey]applicationReference)}
	return p.populate(c.Root, world, applicationReference{root: appType, typ: appType}, "", nil)
}

type populationKey struct {
	descriptor *markerDescriptor
	appType    reflect.Type
}

type populator struct {
	families markFamilies
	active   map[populationKey]applicationReference
}

func (p *populator) populate(descriptor *markerDescriptor, value reflect.Value, source applicationReference, tags reflect.StructTag, qualifiers []string) error {
	if err := descriptor.acceptance.match(source.typ); err != nil {
		return fmt.Errorf("%s as %v (%s): %w", source, descriptor.typ, descriptor.mark, err)
	}
	if err := descriptor.populateMetadata(value, source, tags, qualifiers); err != nil {
		return fmt.Errorf("%s as %v: %w", source, descriptor.typ, err)
	}
	if descriptor.role != landmarkRole {
		if err := p.populateLanes(descriptor, value, source); err != nil {
			return err
		}
	}
	if err := descriptor.populateQualities(value, source.typ, qualifiers); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	if err := descriptor.populateBarrels(value, source.typ); err != nil {
		return fmt.Errorf("%s: %w", source, err)
	}
	return nil
}

func (p *populator) populateLanes(descriptor *markerDescriptor, value reflect.Value, source applicationReference) error {
	appType := utilities.ResolvePointerChain(source.typ)
	if appType == nil || appType.Kind() != reflect.Struct {
		return fmt.Errorf("%s: %v requires an application declaration that resolves to a struct, got %v", source, descriptor.typ, source.typ)
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
		child := source.child(field)
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
