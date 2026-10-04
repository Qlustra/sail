package engine

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/capability"
	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
	"github.com/qlustra/sail/internal/tags"
)

type qualityDescriptor struct {
	name            string
	declarationPath []int
	reader          capability.Reader
}

func compileQuality(field reflect.StructField) (qualityDescriptor, error) {
	typ := reflectutils.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !reflectutils.TypeImplements[decl.QualityDeclaration](typ) {
		return qualityDescriptor{}, fmt.Errorf("quality must embed legend.Quality")
	}
	provider := reflect.New(typ).Interface().(decl.QualityDeclaration)
	shape, err := capability.InvokeDeclaration(typ, "quality", func() decl.QualityShape { return decl.DescribeQuality(provider) })
	if err != nil {
		return qualityDescriptor{}, err
	}
	name, err := tags.ReadSingle(field.Tag, "populate", validateSymbol)
	if err != nil {
		return qualityDescriptor{}, err
	}
	reader, err := capability.CompileReader(shape.Value, shape.Capability, name)
	if err != nil {
		return qualityDescriptor{}, err
	}
	path, err := reflectutils.EmbeddedPath(typ, shape.Carrier)
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
	receiver, err := q.reader.Resolve(receivingType)
	if err != nil {
		return fmt.Errorf("required quality: %w", err)
	}
	destination, err := declarationCarrier(value, q.declarationPath)
	if err != nil {
		return err
	}
	result, err := q.reader.Read(receiver)
	if err != nil {
		return err
	}
	decl.FillQuality(destination.Addr().Interface().(decl.QualityDeclaration), result)
	return nil
}
