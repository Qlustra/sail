package declaration

import (
	"reflect"
	"slices"
	"strings"

	"github.com/qlustra/sail/internal/reflectutils"
)

// Reference belongs to one populated marker, not its shared descriptor.
type Reference struct {
	Root reflect.Type
	Type reflect.Type
	Path []int
}

func (m marker) applicationReference() Reference { return m.source }

// ApplicationType returns the declared Application type described by this marker.
// It is nil until the marker has been populated.
func (m marker) ApplicationType() reflect.Type { return m.source.Type }

// ApplicationPath returns a copy of the field-index path from the Application root.
func (m marker) ApplicationPath() []int { return slices.Clone(m.source.Path) }

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

// SetSource receives the concrete carrier selected by the engine.
func SetSource(carrier any, source Reference) {
	carrier.(interface{ setSource(Reference) }).setSource(source)
}

// ReferenceOf bypasses public accessors which a composite may shadow.
func ReferenceOf(carrier any) Reference {
	return carrier.(interface{ applicationReference() Reference }).applicationReference()
}
