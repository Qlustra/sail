package carriers

import (
	"fmt"

	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules"
	"github.com/qlustra/intel/rules/identifier"
	"github.com/qlustra/intel/rules/property"
)

const (
	InvalidFeatureEncounterCode       logic.CheckCode = "invalid_feature_encounter"
	InvalidFeatureSymbolCode          logic.CheckCode = "invalid_feature_symbol"
	FeatureSymbolEvaluationFailedCode logic.CheckCode = "feature_symbol_evaluation_failed"
)

var featureSymbolRule = identifier.Pattern(SymbolPattern)

// Feature records a field's tag-value definition without a separately compiled
// body declaration.
type Feature struct {
	Encounter *property.Encounter
	// Symbol identifies a consumer tag key. NewFeature validates its syntax;
	// the owning declaration rejects local duplicates when the carrier is added.
	Symbol string
}

// NewFeature reads and validates intrinsic metadata without consulting a host.
func NewFeature(encounter *property.Encounter) (*Feature, logic.Check) {
	if encounter == nil || encounter.Property == nil {
		return nil, logic.Nope(InvalidFeatureEncounterCode, "Feature requires a property encounter")
	}
	symbol := encounter.Property.Tags().Raw("symbol")
	if symbol == "" {
		return nil, logic.Nope(InvalidFeatureSymbolCode, fmt.Sprintf("field %s requires a nonempty symbol", encounter.Property.FieldName()))
	}
	validation := rules.NewValidation("feature symbol")
	matched, valid := featureSymbolRule.ValidateRule(validation, symbol)
	if !valid {
		return nil, logic.Nope(FeatureSymbolEvaluationFailedCode, validation.Trace())
	}
	if !matched {
		return nil, logic.Nope(InvalidFeatureSymbolCode, fmt.Sprintf("field %s: symbol %q must match %s", encounter.Property.FieldName(), symbol, SymbolPattern))
	}
	return &Feature{Encounter: encounter, Symbol: symbol}, logic.Ok()
}

func (f *Feature) String() string {
	if f == nil {
		return "feature <nil>"
	}
	label := fmt.Sprintf("feature %q", f.Symbol)
	if f.Encounter != nil && f.Encounter.Property != nil {
		label += fmt.Sprintf(" (field %s)", f.Encounter.Property.FieldName())
	}
	return label
}

func (f *Feature) Trace() string {
	if f == nil {
		return f.String()
	}
	// A standalone carrier is the sole node; enclosing graph traces assign their
	// own discovery-order identities.
	return f.String() + " #1"
}
