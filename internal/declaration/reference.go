package declaration

import (
	"reflect"
	"slices"
	"strings"

	"github.com/qlustra/sail/internal/reflectutils"
)

// Reference identifies a carrier in a Subject declaration. It belongs to one
// populated marker, not its shared descriptor.
type Reference struct {
	Root reflect.Type // Subject root type.
	Type reflect.Type // Original declared carrier type.
	Path []int        // Field-index path from the Subject root to the carrier.
}

func (m marker) carrierReference() Reference { return m.source }

// CarrierType returns the original declared type of the Subject field represented
// by this marker, or the Subject root type for the root Legend. It is nil until
// the marker has been populated. Pointer types are not normalized.
func (m marker) CarrierType() reflect.Type { return m.source.Type }

// CarrierPath returns a copy of the field-index path from the Subject root to
// this marker's carrier. The root Legend has an empty path.
func (m marker) CarrierPath() []int { return slices.Clone(m.source.Path) }

func (r Reference) Child(field reflect.StructField) Reference {
	return Reference{Root: r.Root, Type: field.Type, Path: slices.Concat(r.Path, field.Index)}
}

func (r Reference) String() string {
	var result strings.Builder
	result.WriteString(r.Root.String())
	typ := r.Root
	for _, index := range r.Path {
		field := reflectutils.ResolvePointerChain(typ).Field(index)
		result.WriteByte('.')
		result.WriteString(field.Name)
		typ = field.Type
	}
	return result.String()
}

// SetSource records a carrier reference in the embedded declaration base
// selected by the engine.
func SetSource(base any, source Reference) {
	base.(interface{ setSource(Reference) }).setSource(source)
}

// ReferenceOf bypasses public accessors which a composite may shadow.
func ReferenceOf(base any) Reference {
	return base.(interface{ carrierReference() Reference }).carrierReference()
}
