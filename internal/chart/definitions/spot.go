package definitions

import (
	"fmt"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules/property"
	"github.com/qlustra/sail/internal/chart/declarations"
)

const SpotKindMismatchCode logic.CheckCode = "spot_kind_mismatch"

func spotTarget(encounter property.Encounter) (*capture.StructInterrogator, logic.Check) {
	return encounter.Property.AsStruct()
}

func spotSliceTarget(encounter property.Encounter) (*capture.StructInterrogator, logic.Check) {
	slice, check := encounter.Property.AsSlice()
	if !check.OK() {
		return nil, check
	}
	structSlice, check := slice.AsStructSlice()
	if !check.OK() {
		return nil, check
	}
	return structSlice.Element(), check
}

func spotPolicy(interrogator *capture.PropertyInterrogator, expected declarations.SpotKind) (declarations.SpotKind, bool, logic.Check) {
	tag, check := interrogator.Tags().Parse[declarations.SpotTag]("spot")
	if !check.OK() {
		return "", false, check
	}
	kind := tag.Format.Get()
	if kind != expected {
		return "", false, logic.Nope(SpotKindMismatchCode, fmt.Sprintf("field %s requires spot:%q, got %q", interrogator.FieldName(), expected, kind))
	}
	return kind, tag.Format.Required(), logic.Ok()
}

func regionSpot(carrier *define.Carrier[declarations.Region], expected declarations.SpotKind) (*declarations.Spot, logic.Check) {
	kind, required, check := spotPolicy(carrier.Property, expected)
	if !check.OK() {
		return nil, check
	}
	return &declarations.Spot{Kind: kind, Required: required, For: declarations.SpotForRegion, Region: carrier}, logic.Ok()
}

func landmarkSpot(carrier *define.Carrier[declarations.Landmark], expected declarations.SpotKind) (*declarations.Spot, logic.Check) {
	kind, required, check := spotPolicy(carrier.Property, expected)
	if !check.OK() {
		return nil, check
	}
	return &declarations.Spot{Kind: kind, Required: required, For: declarations.SpotForLandmark, Landmark: carrier}, logic.Ok()
}
