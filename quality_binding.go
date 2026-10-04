package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

type qualityDescriptor struct {
	name            string
	declarationPath []int
	reader          capabilityReader
}

func compileQuality(field reflect.StructField) (qualityDescriptor, error) {
	typ := utilities.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !utilities.TypeImplements[qualityDeclaration](typ) {
		return qualityDescriptor{}, fmt.Errorf("quality must embed sail.Quality")
	}
	provider := reflect.New(typ).Interface().(qualityDeclaration)
	shape, err := utilities.InvokeDeclaration(typ, "quality", provider.quality)
	if err != nil {
		return qualityDescriptor{}, err
	}
	name, err := readSingleTag(field.Tag, "populate", validateSymbol)
	if err != nil {
		return qualityDescriptor{}, err
	}
	reader, err := compileCapabilityReader(shape.value, shape.capability, name)
	if err != nil {
		return qualityDescriptor{}, err
	}
	path, err := carrierPath(typ, shape.carrier)
	if err != nil {
		return qualityDescriptor{}, err
	}
	return qualityDescriptor{
		name:            field.Name,
		declarationPath: slices.Concat(field.Index, path),
		reader:          reader,
	}, nil
}

func (q qualityDescriptor) populate(value reflect.Value, receivingType reflect.Type) error {
	receiver, err := q.reader.resolve(receivingType)
	if err != nil {
		return fmt.Errorf("required quality: %w", err)
	}
	destination, err := declarationCarrier(value, q.declarationPath)
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
