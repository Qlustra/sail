package values

import (
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/logic"
)

type barrelWriter interface {
	fill(any) logic.Check
	ready() logic.Check
	captureBarrel(*capture.MethodInterrogator, *capture.ResultInterrogator, *capture.TypeInterrogator) logic.Check
}

// Inspection retains provider metadata for one Barrel occurrence.
type Inspection struct {
	method      *capture.MethodInterrogator
	result      *capture.ResultInterrogator
	carrierType *capture.TypeInterrogator
}

func (i Inspection) Method() (*capture.MethodInterrogator, logic.Check) {
	if i.method == nil {
		return nil, logic.Nope("barrel_method_unavailable", "no provider method was captured")
	}
	return i.method, logic.Ok()
}

func (i Inspection) Result() (*capture.ResultInterrogator, logic.Check) {
	if i.result == nil {
		return nil, logic.Nope("barrel_result_unavailable", "no provider result was captured")
	}
	return i.result, logic.Ok()
}

func (i Inspection) CarrierType() (*capture.TypeInterrogator, logic.Check) {
	if i.carrierType == nil {
		return nil, logic.Nope("barrel_carrier_unavailable", "no carrier type was captured")
	}
	return i.carrierType, logic.Ok()
}

func (i *Inspection) captureBarrel(method *capture.MethodInterrogator, result *capture.ResultInterrogator, carrier *capture.TypeInterrogator) logic.Check {
	if i == nil {
		return logic.Nope("invalid_destination", "nil Barrel inspection state")
	}
	if method == nil || result == nil || carrier == nil {
		return logic.Nope("invalid_capture", "incomplete Barrel capture")
	}
	i.method, i.result, i.carrierType = method, result, carrier
	return logic.Ok()
}

func (i *Inspection) ready() logic.Check {
	if i == nil {
		return logic.Nope("invalid_destination", "nil Barrel inspection state")
	}
	return logic.Ok()
}

// CaptureBarrel validates the writable destination before retaining the value
// and inspection state together.
func CaptureBarrel(destination, value any, method *capture.MethodInterrogator, result *capture.ResultInterrogator, carrier *capture.TypeInterrogator) (check logic.Check) {
	defer func() {
		if recover() != nil {
			check = logic.Nope("invalid_destination", "unwritable destination")
		}
	}()
	w, check := barrelDestination(destination)
	if !check.OK() {
		return check
	}
	if method == nil || result == nil || carrier == nil {
		return logic.Nope("invalid_capture", "incomplete Barrel capture")
	}
	if check := w.fill(value); !check.OK() {
		return check
	}
	return w.captureBarrel(method, result, carrier)
}

// ValidateBarrelDestination checks writable occurrence state without mutation.
func ValidateBarrelDestination(destination any) (check logic.Check) {
	defer func() {
		if recover() != nil {
			check = logic.Nope("invalid_destination", "unwritable destination")
		}
	}()
	_, check = barrelDestination(destination)
	return check
}

func barrelDestination(destination any) (barrelWriter, logic.Check) {
	if inspect.IsNil(destination) {
		return nil, logic.Nope("invalid_destination", "nil destination")
	}
	w, ok := destination.(barrelWriter)
	if !ok {
		return nil, logic.Nope("invalid_destination", "destination is not writable Barrel state")
	}
	if check := w.ready(); !check.OK() {
		return nil, check
	}
	return w, logic.Ok()
}
