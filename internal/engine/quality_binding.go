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
	path, err := reflectutils.EmbeddedPath(typ, shape.Base)
	if err != nil {
		return qualityDescriptor{}, err
	}
	return qualityDescriptor{
		name:            field.Name,
		declarationPath: slices.Concat(field.Index, path),
		reader:          reader,
	}, nil
}

func (q qualityDescriptor) populate(value reflect.Value, carrierType reflect.Type) error {
	resolvedType, err := q.reader.Resolve(carrierType)
	if err != nil {
		return fmt.Errorf("required quality: %w", err)
	}
	destination, err := declarationValue(value, q.declarationPath)
	if err != nil {
		return err
	}
	result, err := q.reader.Read(resolvedType)
	if err != nil {
		return err
	}
	decl.FillQuality(destination.Addr().Interface().(decl.QualityDeclaration), result)
	return nil
}
