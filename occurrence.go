package sail

import (
	"reflect"
	"slices"
	"strings"

	"github.com/qlustra/sail/internals/utilities"
)

// subjectReference belongs to a populated marker, not its shared location.
// The root type and copied field path retain identity for later live binding.
type subjectReference struct {
	root reflect.Type
	typ  reflect.Type
	path []int
}

// SubjectType returns the receiving type of this populated marker's occurrence.
// It is nil on a declaration that has not been populated.
func (m marker) SubjectType() reflect.Type { return m.source.typ }

// SubjectPath returns a copy of the field-index path from the subject root.
func (m marker) SubjectPath() []int { return slices.Clone(m.source.path) }

func (r subjectReference) child(field reflect.StructField) subjectReference {
	return subjectReference{root: r.root, typ: field.Type, path: slices.Concat(r.path, field.Index)}
}

func (r subjectReference) String() string {
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

func (l *location) populateMetadata(value reflect.Value, source subjectReference, tags reflect.StructTag, qualifiers []string) error {
	carrier, err := declarationCarrier(value, l.carrierField)
	if err != nil {
		return err
	}
	carrier.Addr().Interface().(interface{ setSource(subjectReference) }).setSource(source)
	for _, feature := range l.features {
		format := SingleValueTag{Tag: Tag{Name: feature.symbol}}
		if _, err := readTag(tags, feature.symbol, &format); err != nil {
			return err
		}
		field, err := declarationCarrier(value, feature.schemaField)
		if err != nil {
			return err
		}
		field.Addr().Interface().(*Feature).Set(format.Raw)
	}
	for _, qualifier := range l.qualifiers {
		field, err := declarationCarrier(value, qualifier.schemaField)
		if err != nil {
			return err
		}
		if slices.Contains(qualifiers, qualifier.symbol) {
			field.Addr().Interface().(*Qualifier).Set()
		}
	}
	return nil
}
