package exploration

import (
	"fmt"

	"github.com/qlustra/intel/compile"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/definitions"
)

type Explorer struct {
	legendDefinition *definitions.LegendDefinition
}

func New() (*Explorer, error) {
	definition, err := define.Struct[definitions.LegendDefinition, declarations.Legend](nil)
	if err != nil {
		return nil, err
	}
	return &Explorer{legendDefinition: definition}, nil
}

func (e *Explorer) Explore[L any]() (*declarations.Legend, error) {
	compiler := compile.Struct[L](e.legendDefinition)
	if !compiler.IsOkay() {
		return nil, fmt.Errorf("cannot compile legend: %s", compiler.Trace())
	}
	legendDeclaration, check := compiler.GetDeclaredValue()
	if !check.OK() {
		return nil, fmt.Errorf("legend not declared: %s", check.Why())
	}
	if check := declarations.ValidateLaneExploration(legendDeclaration); !check.OK() {
		return nil, fmt.Errorf("invalid lane exploration: %s", check.Why())
	}
	vocabulary, check := declarations.BuildVocabulary(legendDeclaration)
	if !check.OK() {
		return nil, fmt.Errorf("invalid qualifier/discriminator namespace: %s", check.Why())
	}
	warnings, check := declarations.AnalyzeLaneConsumption(legendDeclaration)
	if !check.OK() {
		return nil, fmt.Errorf("invalid lane consumption: %s", check.Why())
	}
	legendDeclaration.Warnings = warnings
	legendDeclaration.Vocabulary = vocabulary
	return legendDeclaration, nil
}
