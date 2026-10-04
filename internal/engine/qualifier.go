package engine

import (
	"fmt"
	"reflect"
	"slices"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
	"github.com/qlustra/sail/internal/tags"
)

type qualifierDescriptor struct {
	name            string
	symbol          string
	declarationPath []int
	qualities       []qualityDescriptor
}

func (d *markerDescriptor) collectQualifier(field reflect.StructField, seen map[string]bool) error {
	declaration, err := compileQualifier(field)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", d.typ, field.Name, err)
	}
	if seen[declaration.symbol] {
		return fmt.Errorf("%v.%s: duplicate symbol %q", d.typ, field.Name, declaration.symbol)
	}
	seen[declaration.symbol] = true
	d.qualifiers = append(d.qualifiers, declaration)
	return nil
}

func compileQualifier(field reflect.StructField) (qualifierDescriptor, error) {
	typ := reflectutils.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !reflectutils.TypeImplements[decl.QualifierDeclaration](typ) {
		return qualifierDescriptor{}, fmt.Errorf("qualifier must embed legend.Qualifier")
	}
	symbol, err := tags.ReadSingle(field.Tag, "symbol", validateSymbol)
	if err != nil {
		return qualifierDescriptor{}, err
	}
	path, err := reflectutils.EmbeddedPath(typ, reflect.TypeFor[decl.Qualifier]())
	if err != nil {
		return qualifierDescriptor{}, err
	}
	qualities, err := collectQualities(typ, field.Index, path)
	if err != nil {
		return qualifierDescriptor{}, err
	}
	return qualifierDescriptor{
		name:            field.Name,
		symbol:          symbol,
		declarationPath: slices.Concat(field.Index, path),
		qualities:       qualities,
	}, nil
}

// collectQualities follows the chosen qualifier carrier's embedding path, so
// both a reusable qualifier and wrappers around it can declare requirements.
func collectQualities(typ reflect.Type, prefix, carrier []int) ([]qualityDescriptor, error) {
	var result []qualityDescriptor
	for {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag, tagged, err := readSailTag(field.Tag)
			if err != nil {
				return nil, fmt.Errorf("%v.%s: %w", typ, field.Name, err)
			}
			if !tagged {
				continue
			}
			if !field.IsExported() || tag.Parts[0] != "quality" {
				return nil, fmt.Errorf("%v.%s: expected an exported quality declaration inside a qualifier", typ, field.Name)
			}
			declaration, err := compileQuality(field)
			if err != nil {
				return nil, fmt.Errorf("%v.%s: %w", typ, field.Name, err)
			}
			declaration.declarationPath = slices.Concat(prefix, declaration.declarationPath)
			result = append(result, declaration)
		}
		if len(carrier) == 0 {
			return result, nil
		}
		field := typ.Field(carrier[0])
		typ = reflectutils.ResolvePointerChain(field.Type)
		prefix = slices.Concat(prefix, field.Index)
		carrier = carrier[1:]
	}
}

func (d *markerDescriptor) populateQualities(value reflect.Value, receivingType reflect.Type, symbols []string) error {
	for _, qualifier := range d.qualifiers {
		if !slices.Contains(symbols, qualifier.symbol) {
			continue
		}
		for _, quality := range qualifier.qualities {
			if err := quality.populate(value, receivingType); err != nil {
				return fmt.Errorf("%v.%s.%s (qualifier %q): %w", d.typ, qualifier.name, quality.name, qualifier.symbol, err)
			}
		}
	}
	return nil
}
