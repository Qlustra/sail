package carriers

import (
	"fmt"

	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules/property"
	"github.com/qlustra/sail/internal/chart/behavior"
)

const (
	InvalidBarrelTagCode            logic.CheckCode = "invalid_barrel_tag"
	InvalidBarrelEncounterCode      logic.CheckCode = "invalid_barrel_encounter"
	InvalidBarrelCapabilityCode     logic.CheckCode = "invalid_barrel_capability"
	BarrelShapeEvaluationFailedCode logic.CheckCode = "barrel_shape_evaluation_failed"
)

// Barrel retains a fixed-value field. Required concerns capability availability,
// not whether a successfully captured result is nonzero.
type Barrel struct {
	*ValueBinding
	Required bool
}

// NewBarrel validates Barrel metadata and its fixed-value method shape.
func NewBarrel(encounter *property.Encounter) (*Barrel, logic.Check) {
	if encounter == nil || encounter.Property == nil {
		return nil, logic.Nope(InvalidBarrelEncounterCode, "Barrel requires a property encounter")
	}
	field := encounter.Property.FieldName()
	tags := encounter.Property.Tags()
	if tags.Raw("sail") != "barrel" || tags.Raw("populate") == "" {
		return nil, logic.Nope(InvalidBarrelTagCode, fmt.Sprintf("field %s requires sail:barrel and a nonempty populate tag", field))
	}
	raw, present := tags.Lookup("barrel")
	if present && raw != "optional" && raw != "required" {
		return nil, logic.Nope(InvalidBarrelTagCode, fmt.Sprintf("field %s: barrel tag must be optional or required, got %q", field, raw))
	}
	subject, check := encounter.Property.AsStruct()
	if !check.OK() {
		return nil, check
	}
	shape, check := subject.CreateValue().Invoke("barrelShape", func(value behavior.Barrel) (behavior.ValueShape, logic.Check) {
		return behavior.BarrelShape(value), logic.Ok()
	})
	if !check.OK() {
		return nil, logic.NopeBecause(BarrelShapeEvaluationFailedCode, fmt.Sprintf("field %s barrel shape", field), check)
	}
	binding, check := newValueBinding(encounter, shape)
	if !check.OK() {
		return nil, logic.NopeBecause(InvalidBarrelCapabilityCode, fmt.Sprintf("field %s Barrel binding", field), check)
	}
	return &Barrel{ValueBinding: binding, Required: present && raw == "required"}, logic.Ok()
}
