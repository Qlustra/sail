package populating

import (
	"fmt"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules"
	"github.com/qlustra/intel/rules/typerule"
	"github.com/qlustra/sail/internal/chart/declarations"
)

// AcceptLandmark evaluates a Landmark's acceptance rule against the original
// captured type, without narrowing or mutating shared declaration state.
func AcceptLandmark(landmark *declarations.Landmark, subject *capture.TypeInterrogator) logic.Check {
	if landmark == nil {
		return logic.Nope("invalid_landmark_acceptance", "Landmark is nil")
	}
	if subject == nil {
		return logic.Nope("invalid_landmark_acceptance", "subject type is nil")
	}
	if landmark.Acceptance == nil {
		return logic.Ok()
	}

	validation := rules.NewValidation("landmark acceptance")
	matched, valid := landmark.Acceptance.ValidateRule(validation, typerule.Encounter{
		EncounterContext: &rules.EncounterContext{},
		Type:             subject.CapturedType(),
	})
	if !valid {
		return logic.Nope("landmark_acceptance_evaluation_failed", validation.Trace())
	}
	if !matched {
		return logic.Nope("landmark_not_accepted", fmt.Sprintf("%s does not satisfy Landmark acceptance: %s", subject, validation.Trace()))
	}
	return logic.Ok()
}
