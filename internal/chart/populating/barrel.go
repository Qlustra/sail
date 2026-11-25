package populating

import (
	"fmt"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/sail/internal/chart/carriers"
	"github.com/qlustra/sail/internal/chart/values"
)

// CaptureBarrel obtains an occurrence's value from a fresh zero-valued
// consumer carrier. Capability absence is optional only when the compiled
// Barrel says it is optional; invocation and destination failures are errors.
func CaptureBarrel(b *carriers.Barrel, carrierType *capture.TypeInterrogator, destination any) logic.Check {
	field := "<unknown>"
	if b != nil && b.ValueBinding != nil && b.Encounter != nil && b.Encounter.Property != nil {
		field = b.Encounter.Property.FieldName()
	}
	configFailure := func(message string) logic.Check {
		return logic.Nope("invalid_barrel_capture", fmt.Sprintf("field %s: %s", field, message))
	}
	if b == nil || b.ValueBinding == nil {
		return configFailure("missing compiled Barrel")
	}
	binding := b.ValueBinding
	if binding.Encounter == nil || binding.Encounter.Property == nil || binding.Populate == "" || binding.Shape.Result == nil || binding.Shape.Result.TypeInspector == nil || binding.Shape.Capability == nil || binding.Shape.Capability.TypeInspector == nil {
		return configFailure("incomplete compiled Barrel metadata")
	}
	resultType, capabilityType := binding.Shape.Result.CapturedType(), binding.Shape.Capability.CapturedType()
	if resultType == nil || capabilityType == nil {
		return configFailure("Barrel result or capability type is nil")
	}
	capability, check := binding.Shape.Capability.AsInterface()
	if !check.OK() {
		return configFailure("Barrel capability is not an interface")
	}
	if !capability.IsEmpty() && binding.Method == nil {
		return configFailure("literal Barrel capability has no compiled method")
	}
	if carrierType == nil || carrierType.TypeInspector == nil || carrierType.CapturedType() == nil {
		return configFailure("missing consumer carrier type")
	}
	if check := values.ValidateBarrelDestination(destination); !check.OK() {
		return logic.NopeBecause("invalid_barrel_destination", fmt.Sprintf("field %s destination", field), check)
	}
	unavailable := func(cause logic.Check) logic.Check {
		if !b.Required {
			return logic.Ok()
		}
		return logic.NopeBecause("barrel_unavailable", fmt.Sprintf("field %s requires capability %q", field, binding.Populate), cause)
	}
	receiver, check := carrierType.ZeroReceiver()
	if !check.OK() {
		return unavailable(check)
	}
	receiverType := receiver.Type()
	if receiverType == nil {
		return unavailable(logic.Nope("barrel_receiver_unavailable", "zero receiver has no type"))
	}
	if check := receiverType.Satisfies(capabilityType); !check.OK() {
		return unavailable(check)
	}
	method, check := receiver.Method(binding.Populate)
	if !check.OK() {
		return unavailable(check)
	}
	if method.InputArity() != 0 || method.OutputArity() != 1 || !method.ResultAssignableTo(resultType) {
		return unavailable(logic.Nope("barrel_method_signature_invalid", fmt.Sprintf("%s must take no inputs and return one assignable result", binding.Populate)))
	}
	getter, check := receiver.Getter(method)
	if !check.OK() {
		return unavailable(check)
	}
	result, check := getter.Read()
	if !check.OK() {
		return logic.NopeBecause("barrel_invocation_failed", fmt.Sprintf("field %s method %s", field, binding.Populate), check)
	}
	if check := values.CaptureBarrel(destination, result.Get(), method, result, carrierType); !check.OK() {
		return logic.NopeBecause("barrel_write_failed", fmt.Sprintf("field %s", field), check)
	}
	return logic.Ok()
}

// captureBarrels fills either a direct occurrence body or an activated Qualifier
// body. Both use the original Chart carrier, never the Legend destination type.
func captureBarrels(barrelCarriers []*carriers.Barrel, recipeCarrierType *capture.TypeInterrogator, legendOccurrence *capture.StructValueInterrogator) error {
	for _, barrelCarrier := range barrelCarriers {
		legendBarrelProperty := barrelCarrier.Encounter.Property
		legendBarrelValue, check := legendOccurrence.Field(legendBarrelProperty.FieldDescriptor)
		if !check.OK() {
			return fmt.Errorf("Barrel %s destination: %s", legendBarrelProperty.FieldName(), check.Why())
		}
		legendBarrelAddress, check := legendBarrelValue.Address()
		if !check.OK() {
			return fmt.Errorf("Barrel %s address: %s", legendBarrelProperty.FieldName(), check.Why())
		}
		if check := CaptureBarrel(barrelCarrier, recipeCarrierType, legendBarrelAddress); !check.OK() {
			return fmt.Errorf("Barrel %s: %s", legendBarrelProperty.FieldName(), check.Why())
		}
	}
	return nil
}
