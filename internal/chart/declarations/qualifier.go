package declarations

import (
	"fmt"
	"strings"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules"
	"github.com/qlustra/intel/rules/identifier"
	"github.com/qlustra/intel/rules/property"
	"github.com/qlustra/sail/internal/chart/carriers"
)

// QualifierCarrier retains one symbol-bearing field use. Declaration is nil for
// the bare presence-only form; custom uses can share a compiled body.
type QualifierCarrier struct {
	Property    *capture.PropertyInterrogator
	Symbol      string
	Declaration *Qualifier
}

func NewQualifierCarrier(encounter *property.Encounter, declaration *Qualifier) (*QualifierCarrier, logic.Check) {
	if encounter == nil || encounter.Property == nil {
		return nil, logic.Nope("invalid_qualifier_encounter", "Qualifier requires a property encounter")
	}
	symbol := encounter.Property.Tags().Raw("symbol")
	validation := rules.NewValidation("qualifier symbol")
	matched, valid := qualifierSymbolRule.ValidateRule(validation, symbol)
	if !valid {
		return nil, logic.Nope("qualifier_symbol_evaluation_failed", validation.Trace())
	}
	if !matched {
		return nil, logic.Nope("invalid_qualifier_symbol", fmt.Sprintf("field %s: symbol %q must match %s", encounter.Property.FieldName(), symbol, carriers.SymbolPattern))
	}
	return &QualifierCarrier{Property: encounter.Property, Symbol: symbol, Declaration: declaration}, logic.Ok()
}

var qualifierSymbolRule = identifier.Pattern(carriers.SymbolPattern)

const (
	MissingQualifierCode         logic.CheckCode = "missing_qualifier"
	DuplicateQualifierSymbolCode logic.CheckCode = "duplicate_qualifier_symbol"
)

func addQualifier(qualifiers []*QualifierCarrier, qualifier *QualifierCarrier) ([]*QualifierCarrier, logic.Check) {
	if qualifier == nil {
		return qualifiers, logic.Nope(MissingQualifierCode, "cannot add a nil Qualifier")
	}
	for _, existing := range qualifiers {
		if existing != nil && existing.Symbol == qualifier.Symbol {
			return qualifiers, logic.Nope(DuplicateQualifierSymbolCode, fmt.Sprintf("fields %q and %q have duplicate qualifier symbol %q", existing.Property.FieldName(), qualifier.Property.FieldName(), qualifier.Symbol))
		}
	}
	return append(qualifiers, qualifier), logic.Ok()
}

func (q *QualifierCarrier) Trace() string {
	t := declarationTrace{seen: make(map[any]int), active: make(map[any]bool)}
	t.walkQualifier(q, 0)
	return strings.TrimSuffix(t.output.String(), "\n")
}
