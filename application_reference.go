package sail

import (
	"reflect"
	"slices"
	"strings"

	"github.com/qlustra/sail/internals/utilities"
)

// applicationReference belongs to a populated marker, not its shared descriptor.
// The root type and copied field path retain identity for later live binding.
type applicationReference struct {
	root reflect.Type
	typ  reflect.Type
	path []int
}

func (m marker) applicationReference() applicationReference { return m.source }

// ApplicationType returns the declared application type described by this marker.
// It is nil until the marker has been populated.
func (m marker) ApplicationType() reflect.Type { return m.source.typ }

// ApplicationPath returns a copy of the field-index path from the application root.
func (m marker) ApplicationPath() []int { return slices.Clone(m.source.path) }

func (r applicationReference) child(field reflect.StructField) applicationReference {
	return applicationReference{root: r.root, typ: field.Type, path: slices.Concat(r.path, field.Index)}
}

func (r applicationReference) String() string {
	var result strings.Builder
	result.WriteString(r.root.String())
	typ := r.root
	for _, index := range r.path {
		field := utilities.ResolvePointerChain(typ).Field(index)
		result.WriteByte('.')
		result.WriteString(field.Name)
		typ = field.Type
	}
	return result.String()
}

func (d *markerDescriptor) populateMetadata(value reflect.Value, source applicationReference, tags reflect.StructTag, qualifiers []string) error {
	carrier, err := declarationCarrier(value, d.carrierPath)
	if err != nil {
		return err
	}
	carrier.Addr().Interface().(interface{ setSource(applicationReference) }).setSource(source)
	for _, feature := range d.features {
		format := SingleValueTag{Tag: Tag{Name: feature.symbol}}
		if _, err := readTag(tags, feature.symbol, &format); err != nil {
			return err
		}
		field, err := declarationCarrier(value, feature.declarationPath)
		if err != nil {
			return err
		}
		field.Addr().Interface().(*Feature).Set(format.Raw)
	}
	for _, qualifier := range d.qualifiers {
		field, err := declarationCarrier(value, qualifier.declarationPath)
		if err != nil {
			return err
		}
		if slices.Contains(qualifiers, qualifier.symbol) {
			field.Addr().Interface().(*Qualifier).Set()
		}
	}
	return nil
}
