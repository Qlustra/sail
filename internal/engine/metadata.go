package engine

import (
	"reflect"
	"slices"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/tags"
)

func (d *markerDescriptor) populateMetadata(value reflect.Value, source decl.Reference, structTags reflect.StructTag, qualifiers []string) error {
	carrier, err := declarationCarrier(value, d.carrierPath)
	if err != nil {
		return err
	}
	decl.SetSource(carrier.Addr().Interface(), source)
	for _, feature := range d.features {
		format := tags.SingleValue{Tag: tags.Tag{Name: feature.symbol}}
		if _, err := tags.Read(structTags, feature.symbol, &format); err != nil {
			return err
		}
		field, err := declarationCarrier(value, feature.declarationPath)
		if err != nil {
			return err
		}
		field.Addr().Interface().(*decl.Feature).Set(format.Raw)
	}
	for _, qualifier := range d.qualifiers {
		field, err := declarationCarrier(value, qualifier.declarationPath)
		if err != nil {
			return err
		}
		if slices.Contains(qualifiers, qualifier.symbol) {
			field.Addr().Interface().(*decl.Qualifier).Set()
		}
	}
	return nil
}
