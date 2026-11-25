package carriers

import (
	"fmt"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules/property"
	"github.com/qlustra/sail/internal/chart/behavior"
)

// ValueBinding is a compiled field-to-capability instruction, not a captured
// consumer value. Availability belongs to Barrel and qualifier activation.
// Dependent result relationships are Framework use-time assertions; interface
// contracts are literal, not signature templates.
type ValueBinding struct {
	Encounter *property.Encounter
	Shape     behavior.ValueShape
	Populate  string
	Method    *capture.InterfaceMethodInterrogator // nil for an empty-interface dynamic binding
}

// The Barrel constructor checks its encounter, tag and shape hook first.
// This step prepares the instruction without invoking a provider. An empty
// interface leaves the actual getter signature to be checked when binding a
// Chart carrier; a nonempty interface supplies a static method prototype.
func newValueBinding(encounter *property.Encounter, shape behavior.ValueShape) (*ValueBinding, logic.Check) {
	field := encounter.Property.FieldName()
	if shape.Capability == nil || shape.Result == nil {
		return nil, logic.Nope("invalid_value_binding", fmt.Sprintf("field %s returned an incomplete value shape", field))
	}
	capability, check := shape.Capability.AsInterface()
	if !check.OK() {
		return nil, logic.NopeBecause("invalid_value_capability", fmt.Sprintf("field %s requires an interface capability", field), check)
	}
	populate := encounter.Property.Tags().Raw("populate")
	if !inspect.IsExportedIdentifier(populate) {
		return nil, logic.Nope("invalid_value_capability", fmt.Sprintf("field %s populate method %q must be an exported Go identifier", field, populate))
	}
	if capability.IsEmpty() {
		return &ValueBinding{Encounter: encounter, Shape: shape, Populate: populate}, logic.Ok()
	}
	method, check := capability.Method(populate)
	if !check.OK() {
		return nil, logic.NopeBecause("invalid_value_capability", fmt.Sprintf("field %s populate method %q", field, populate), check)
	}
	if method.InputArity() != 0 || method.OutputArity() != 1 || !method.ResultAssignableTo(shape.Result.CapturedType()) {
		return nil, logic.Nope("invalid_value_capability", fmt.Sprintf("field %s populate method %q must take no inputs and return one result assignable to the declared value type", field, populate))
	}
	return &ValueBinding{Encounter: encounter, Shape: shape, Populate: populate, Method: method}, logic.Ok()
}
