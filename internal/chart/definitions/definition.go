package definitions

import (
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/options"
	"github.com/qlustra/intel/rules/extension"
	"github.com/qlustra/intel/rules/field"
	"github.com/qlustra/intel/rules/property"
	"github.com/qlustra/sail/internal/chart/behavior"
)

type MarkerDefinition[HOST define.ScopeDeclaration] struct {
	define.StructDefinition[HOST]
	MarkerType capture.StructInterrogator
}

func (d *MarkerDefinition[HOST]) Marker[M behavior.Marker](marker string, validators ...extension.RuleAPI) *define.StructExtensionDefinition {
	extensionDefinition := d.Extension(
		options.Required,
		extension.Satisfies[M](),
		extension.Field(field.TaggedAsSingleValue("sail", marker)),
	)
	extensionDefinition.Also(validators...)
	return extensionDefinition
}

func (d *MarkerDefinition[HOST]) Carrier[
	INDICATOR define.LeafIndicator,
](validators ...property.RuleAPI) *CarrierDefinition[HOST, INDICATOR] {
	carrier := Carrier[
		HOST,
		INDICATOR,
	](
		define.DeclarationKey,
	)
	d.Property(options.Optional, carrier.Indicator(), validators...).Encounter(carrier.Handle)
	return carrier
}
