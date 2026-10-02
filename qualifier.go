package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

type qualifier struct {
	name        string
	symbol      string
	schemaField []int
	qualities   []quality
}

func (l *location) collectQualifier(field reflect.StructField, seen map[string]bool) error {
	declaration, err := compileQualifier(field)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", l.typ, field.Name, err)
	}
	if seen[declaration.symbol] {
		return fmt.Errorf("%v.%s: duplicate symbol %q", l.typ, field.Name, declaration.symbol)
	}
	seen[declaration.symbol] = true
	l.qualifiers = append(l.qualifiers, declaration)
	return nil
}

func compileQualifier(field reflect.StructField) (qualifier, error) {
	typ := utilities.ResolvePointerChain(field.Type)
	if typ == nil || typ.Kind() != reflect.Struct || !utilities.TypeImplements[qualifierDeclaration](typ) {
		return qualifier{}, fmt.Errorf("qualifier must embed sail.Qualifier")
	}
	symbol, err := readSingleTag(field.Tag, "symbol", validateSymbol)
	if err != nil {
		return qualifier{}, err
	}
	path, err := carrierPath(typ, reflect.TypeFor[Qualifier]())
	if err != nil {
		return qualifier{}, err
	}
	qualities, err := collectQualities(typ, field.Index, path)
	if err != nil {
		return qualifier{}, err
	}
	return qualifier{
		name:        field.Name,
		symbol:      symbol,
		schemaField: slices.Concat(field.Index, path),
		qualities:   qualities,
	}, nil
}

// collectQualities follows the chosen qualifier carrier's embedding path, so
// both a reusable qualifier and wrappers around it can declare requirements.
func collectQualities(typ reflect.Type, prefix, carrier []int) ([]quality, error) {
	var result []quality
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
			declaration.schemaField = slices.Concat(prefix, declaration.schemaField)
			result = append(result, declaration)
		}
		if len(carrier) == 0 {
			return result, nil
		}
		field := typ.Field(carrier[0])
		typ = utilities.ResolvePointerChain(field.Type)
		prefix = slices.Concat(prefix, field.Index)
		carrier = carrier[1:]
	}
}

func (l *location) populateQualities(value reflect.Value, subject reflect.Type, symbols []string) error {
	for _, qualifier := range l.qualifiers {
		if !slices.Contains(symbols, qualifier.symbol) {
			continue
		}
		for _, quality := range qualifier.qualities {
			if err := quality.populate(value, subject); err != nil {
				return fmt.Errorf("%v.%s.%s (qualifier %q): %w", l.typ, qualifier.name, quality.name, qualifier.symbol, err)
			}
		}
	}
	return nil
}
