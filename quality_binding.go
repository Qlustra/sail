package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

type quality struct {
	name        string
	schemaField []int
	reader      capabilityReader
}

func compileQuality(field reflect.StructField) (quality, error) {
	typ := utilities.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !utilities.TypeImplements[qualityDeclaration](typ) {
		return quality{}, fmt.Errorf("quality must embed sail.Quality")
	}
	provider := reflect.New(typ).Interface().(qualityDeclaration)
	shape, err := utilities.InvokeDeclaration(typ, "quality", provider.quality)
	if err != nil {
		return quality{}, err
	}
	name, err := readSingleTag(field.Tag, "populate", validateSymbol)
	if err != nil {
		return quality{}, err
	}
	reader, err := compileCapabilityReader(shape.value, shape.capability, name)
	if err != nil {
		return quality{}, err
	}
	path, err := carrierPath(typ, shape.carrier)
	if err != nil {
		return quality{}, err
	}
	return quality{
		name:        field.Name,
		schemaField: slices.Concat(field.Index, path),
		reader:      reader,
	}, nil
}

func (q quality) populate(value reflect.Value, subject reflect.Type) error {
	receiver, err := q.reader.resolve(subject)
	if err != nil {
		return fmt.Errorf("required quality: %w", err)
	}
	destination, err := declarationCarrier(value, q.schemaField)
	if err != nil {
		return err
	}
	result, err := q.reader.read(receiver)
	if err != nil {
		return err
	}
	destination.Addr().Interface().(qualityDeclaration).fillQuality(result)
	return nil
}
