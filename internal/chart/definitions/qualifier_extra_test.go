package definitions

import (
	"reflect"
	"strings"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/legend"
)

var qualifierProviderCalls int

type countedProvider interface{ Value() string }

type countedProviderImpl struct{}

func (countedProviderImpl) Value() string { qualifierProviderCalls++; return "value" }

type customQualifierMarkerBody struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type customQualifierMarkerRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             customQualifierMarkerBody `sail:"qualifier" symbol:"q"`
}

type privateNonQualityQualifierBody struct {
	legend.Qualifier `sail:"qualifier"`
	internal         string
	Value            legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type privateNonQualityQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             privateNonQualityQualifierBody `sail:"qualifier" symbol:"q"`
}

type extraLandmarkBody struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[any, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type extraLandmark struct {
	legend.Landmark `sail:"landmark" family:"test"`
	Q               extraLandmarkBody `sail:"qualifier" symbol:"landmark-value"`
}

type invalidCapabilityBody struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[string, countedProviderImpl] `sail:"barrel" populate:"Value" barrel:"required"`
}

type wrongQualityTagBody struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[string, countedProvider] `sail:"feature" populate:"Value" barrel:"required"`
}

type legacyQualityTagBody struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[string, countedProvider] `sail:"quality" populate:"Value" barrel:"required"`
}

type legacyQualityTagRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             legacyQualityTagBody `sail:"qualifier" symbol:"q"`
}

type missingFieldQualifierBody struct {
	legend.Qualifier
	Value legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type extraRegionQualifierBody struct {
	legend.Qualifier `sail:"qualifier"`
	legend.Region    `sail:"region" family:"extra"`
	Value            legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type extraFeatureQualifierBody struct {
	legend.Qualifier `sail:"qualifier"`
	legend.Feature   `sail:"feature"`
	Value            legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type unrelatedQualifierEmbedded struct{ Marker string }

type unrelatedQualifierInterface interface{ Marker() }

type unrelatedQualifierStructBody struct {
	legend.Qualifier `sail:"qualifier"`
	unrelatedQualifierEmbedded
	Value legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type unrelatedQualifierInterfaceBody struct {
	legend.Qualifier `sail:"qualifier"`
	unrelatedQualifierInterface
	Value legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value" barrel:"required"`
}

type extraRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             extraLandmarkBody `sail:"qualifier" symbol:"q"`
}

type invalidCapabilityRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             invalidCapabilityBody `sail:"qualifier" symbol:"q"`
}

type wrongQualityTagRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             wrongQualityTagBody `sail:"qualifier" symbol:"q"`
}

type missingFieldQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             missingFieldQualifierBody `sail:"qualifier" symbol:"q"`
}

type extraRegionQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             extraRegionQualifierBody `sail:"qualifier" symbol:"q"`
}

type extraFeatureQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             extraFeatureQualifierBody `sail:"qualifier" symbol:"q"`
}

type unrelatedQualifierStructRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             unrelatedQualifierStructBody `sail:"qualifier" symbol:"q"`
}

type unrelatedQualifierInterfaceRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             unrelatedQualifierInterfaceBody `sail:"qualifier" symbol:"q"`
}

type invalidSymbolRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             legend.Qualifier `sail:"qualifier" symbol:""`
}

type cachedSecondBadSymbolRegion struct {
	legend.Region `sail:"region" family:"test"`
	First         testBody `sail:"qualifier" symbol:"good"`
	Second        testBody `sail:"qualifier" symbol:""`
}

type multipleBareQualifiersRegion struct {
	legend.Region `sail:"region" family:"test"`
	First         legend.Qualifier `sail:"qualifier" symbol:"one"`
	Second        legend.Qualifier `sail:"qualifier" symbol:"two"`
}

func TestQualifierExtraValidLandmarkAndAnyResult(t *testing.T) {
	definition, err := define.Struct[LandmarkDefinition, declarations.Landmark](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[extraLandmark]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	qualifierProviderCalls = 0
	ctx, validation := definition.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatalf("compile: %s", validation.Trace())
	}
	body, check := ctx.Access[declarations.Landmark](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(body.Qualifiers) != 1 || body.Qualifiers[0].Symbol != "landmark-value" {
		t.Fatalf("landmark qualifier metadata: %#v", body.Qualifiers)
	}
	if qualifierProviderCalls != 0 {
		t.Fatalf("provider invoked %d times during compilation", qualifierProviderCalls)
	}
	trace := body.Trace()
	for _, want := range []string{"landmark-value", "Value", "barrel required=true", `populate="Value"`, "capability=", "result=interface {}"} {
		if !strings.Contains(trace, want) {
			t.Fatalf("trace missing %q: %s", want, trace)
		}
	}
}

func TestQualifierMarkerInternalFieldsAndInvalidMembers(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  any
		okay bool
	}{
		{"custom marker internal field", customQualifierMarkerRegion{}, true},
		{"empty body", qualifierEmptyRegion{}, false},
		{"direct private non-Quality field", privateNonQualityQualifierRegion{}, false},
		{"other authored embedding", unrelatedQualifierStructRegion{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
			if err != nil {
				t.Fatal(err)
			}
			subject, check := capture.TypeReflection(reflect.TypeOf(tc.typ)).AsStruct()
			if !check.OK() {
				t.Fatal(check.Why())
			}
			_, validation := definition.Compile(subject, nil)
			if got := validation.IsOkay(); got != tc.okay {
				t.Fatalf("validation okay = %v; want %v; trace: %s", got, tc.okay, validation.Trace())
			}
		})
	}
}

func TestQualifierExtraInvalidMetadata(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  any
		want string
	}{
		{"non-interface capability", invalidCapabilityRegion{}, "invalid property"},
		{"wrong quality tag", wrongQualityTagRegion{}, "invalid property"},
		{"legacy quality tag", legacyQualityTagRegion{}, "invalid property"},
		{"missing field qualifier tag", missingFieldQualifierRegion{}, "qualifier"},
		{"extra embedded Region", extraRegionQualifierRegion{}, "Region"},
		{"extra embedded Feature", extraFeatureQualifierRegion{}, "Feature"},
		{"unrelated embedded struct", unrelatedQualifierStructRegion{}, "Q"},
		{"unrelated embedded interface", unrelatedQualifierInterfaceRegion{}, "unrelatedQualifierInterface"},
		{"invalid symbol", invalidSymbolRegion{}, "symbol"},
		{"cached second bad symbol", cachedSecondBadSymbolRegion{}, "Second"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
			if err != nil {
				t.Fatal(err)
			}
			subject, check := capture.TypeReflection(reflect.TypeOf(tc.typ)).AsStruct()
			if !check.OK() {
				t.Fatal(check.Why())
			}
			_, validation := definition.Compile(subject, nil)
			if validation.IsOkay() {
				t.Fatal("invalid metadata compiled successfully")
			}
			if trace := validation.Trace(); !strings.Contains(trace, tc.want) {
				t.Fatalf("trace missing %q: %s", tc.want, trace)
			}
		})
	}
}

func TestQualifierExtraBareMultipleSymbols(t *testing.T) {
	definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[multipleBareQualifiersRegion]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := definition.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatalf("compile multiple bare qualifiers: %s", validation.Trace())
	}
	body, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(body.Qualifiers) != 2 || body.Qualifiers[0].Symbol != "one" || body.Qualifiers[1].Symbol != "two" {
		t.Fatalf("qualifier metadata: %#v", body.Qualifiers)
	}
}

type promotedQualifierFields struct {
	Default legend.Qualifier `sail:"qualifier" symbol:"default"`
}

type promotedQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	promotedQualifierFields
}

func TestBareQualifierRecognitionIsIndependentOfFieldDepth(t *testing.T) {
	definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[promotedQualifierRegion]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := definition.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatal(validation.Trace())
	}
	body, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(body.Qualifiers) != 1 {
		t.Fatalf("qualifier uses = %d; want 1", len(body.Qualifiers))
	}
	use := body.Qualifiers[0]
	if use.Symbol != "default" || use.Declaration != nil || len(use.Property.Index()) != 2 {
		t.Fatalf("promoted presence-only qualifier was not retained: %s", body.Trace())
	}
}
