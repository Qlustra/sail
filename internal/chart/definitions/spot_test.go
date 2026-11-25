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

type occurrenceRegion struct {
	legend.Region `sail:"region" family:"region-family"`
}
type occurrenceLandmark struct {
	legend.Landmark `sail:"landmark" family:"landmark-family"`
}

type directRegionOptional struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single"`
}
type directRegionRequired struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single,required"`
}
type sliceRegionOptional struct {
	legend.Lane
	Member []occurrenceRegion `sail:"spot" spot:"multi"`
}
type sliceRegionRequired struct {
	legend.Lane
	Member []occurrenceRegion `sail:"spot" spot:"multi,required"`
}
type directLandmarkOptional struct {
	legend.Lane
	Member occurrenceLandmark `sail:"spot" spot:"single"`
}
type directLandmarkRequired struct {
	legend.Lane
	Member occurrenceLandmark `sail:"spot" spot:"single,required"`
}
type sliceLandmarkOptional struct {
	legend.Lane
	Member []occurrenceLandmark `sail:"spot" spot:"multi"`
}
type sliceLandmarkRequired struct {
	legend.Lane
	Member []occurrenceLandmark `sail:"spot" spot:"multi,required"`
}

type spotEmpty struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:""`
}
type spotRepeatedQualifier struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single,required,required"`
}
type spotEmptyQualifier struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single,"`
}
type spotUnknownQualifier struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single,optional"`
}
type spotMisplacedRepresentation struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"required,single"`
}
type spotMissingTag struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot"`
}
type malformedMulti struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"multi"`
}
type malformedMultiRepeatedRequired struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"multi,required,required"`
}
type malformedContradictoryRepresentation struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single,multi"`
}
type malformedCaseRepresentation struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"Single"`
}
type malformedLeadingWhitespace struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:" single"`
}
type malformedTrailingWhitespace struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"single "`
}

type directRegionMulti struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"multi"`
}
type directRegionMultiRequired struct {
	legend.Lane
	Member occurrenceRegion `sail:"spot" spot:"multi,required"`
}
type sliceRegionSingle struct {
	legend.Lane
	Member []occurrenceRegion `sail:"spot" spot:"single"`
}
type sliceRegionSingleRequired struct {
	legend.Lane
	Member []occurrenceRegion `sail:"spot" spot:"single,required"`
}
type directLandmarkMulti struct {
	legend.Lane
	Member occurrenceLandmark `sail:"spot" spot:"multi"`
}
type directLandmarkMultiRequired struct {
	legend.Lane
	Member occurrenceLandmark `sail:"spot" spot:"multi,required"`
}
type sliceLandmarkSingle struct {
	legend.Lane
	Member []occurrenceLandmark `sail:"spot" spot:"single"`
}
type sliceLandmarkSingleRequired struct {
	legend.Lane
	Member []occurrenceLandmark `sail:"spot" spot:"single,required"`
}

func TestSpotOccurrenceCompilation(t *testing.T) {
	lane, err := define.Struct[LaneDefinition, declarations.Lane](nil)
	if err != nil {
		t.Fatalf("define lane: %v", err)
	}
	cases := []struct {
		name     string
		typ      reflect.Type
		kind     declarations.SpotKind
		forKind  declarations.SpotFor
		required bool
	}{
		{"direct Region optional", reflect.TypeFor[directRegionOptional](), declarations.SingleSpot, declarations.SpotForRegion, false},
		{"direct Region required", reflect.TypeFor[directRegionRequired](), declarations.SingleSpot, declarations.SpotForRegion, true},
		{"slice Region optional", reflect.TypeFor[sliceRegionOptional](), declarations.MultiSpot, declarations.SpotForRegion, false},
		{"slice Region required", reflect.TypeFor[sliceRegionRequired](), declarations.MultiSpot, declarations.SpotForRegion, true},
		{"direct Landmark optional", reflect.TypeFor[directLandmarkOptional](), declarations.SingleSpot, declarations.SpotForLandmark, false},
		{"direct Landmark required", reflect.TypeFor[directLandmarkRequired](), declarations.SingleSpot, declarations.SpotForLandmark, true},
		{"slice Landmark optional", reflect.TypeFor[sliceLandmarkOptional](), declarations.MultiSpot, declarations.SpotForLandmark, false},
		{"slice Landmark required", reflect.TypeFor[sliceLandmarkRequired](), declarations.MultiSpot, declarations.SpotForLandmark, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subject, check := capture.TypeReflection(tc.typ).AsStruct()
			if !check.OK() {
				t.Fatalf("capture subject: %s", check.Why())
			}
			ctx, validation := lane.Compile(subject, nil)
			if !validation.IsOkay() {
				t.Fatalf("compile: %s", validation.Trace())
			}
			body, check := ctx.Access[declarations.Lane](define.DeclarationKey)
			if !check.OK() {
				t.Fatalf("access Lane declaration: %s", check.Why())
			}
			trace := body.Trace()
			if !strings.Contains(trace, "field Member") {
				t.Fatalf("trace does not retain spot field: %s", trace)
			}
			if !strings.Contains(trace, string(tc.kind)+" "+string(tc.forKind)) {
				t.Fatalf("trace missing representation/target %s %s: %s", tc.kind, tc.forKind, trace)
			}
			if tc.required && !strings.Contains(trace, "(required)") {
				t.Fatalf("required spot missing from trace: %s", trace)
			}
			if !tc.required && strings.Contains(trace, "(required)") {
				t.Fatalf("optional spot marked required: %s", trace)
			}
		})
	}
}

func TestSpotOccurrenceMalformedTags(t *testing.T) {
	lane, err := define.Struct[LaneDefinition, declarations.Lane](nil)
	if err != nil {
		t.Fatalf("define lane: %v", err)
	}
	for _, tc := range []struct {
		name string
		typ  reflect.Type
	}{
		{"empty tag", reflect.TypeFor[spotEmpty]()},
		{"repeated qualifier", reflect.TypeFor[spotRepeatedQualifier]()},
		{"empty qualifier", reflect.TypeFor[spotEmptyQualifier]()},
		{"unknown qualifier", reflect.TypeFor[spotUnknownQualifier]()},
		{"misplaced representation", reflect.TypeFor[spotMisplacedRepresentation]()},
		{"missing tag", reflect.TypeFor[spotMissingTag]()},
		{"multi without slice", reflect.TypeFor[malformedMulti]()},
		{"multi repeated required", reflect.TypeFor[malformedMultiRepeatedRequired]()},
		{"contradictory representation", reflect.TypeFor[malformedContradictoryRepresentation]()},
		{"case mismatch", reflect.TypeFor[malformedCaseRepresentation]()},
		{"leading whitespace", reflect.TypeFor[malformedLeadingWhitespace]()},
		{"trailing whitespace", reflect.TypeFor[malformedTrailingWhitespace]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			subject, check := capture.TypeReflection(tc.typ).AsStruct()
			if !check.OK() {
				t.Fatalf("capture subject: %s", check.Why())
			}
			_, validation := lane.Compile(subject, nil)
			if validation.IsOkay() {
				t.Fatal("compilation succeeded for malformed spot")
			}
			if !strings.Contains(validation.Trace(), "spot") {
				t.Fatalf("failure does not identify spot metadata: %s", validation.Trace())
			}
		})
	}
}

func TestSpotShapeMismatchesReportKindMismatch(t *testing.T) {
	lane, err := define.Struct[LaneDefinition, declarations.Lane](nil)
	if err != nil {
		t.Fatalf("define lane: %v", err)
	}
	for _, tc := range []struct {
		name string
		typ  reflect.Type
	}{
		{"direct Region multi", reflect.TypeFor[directRegionMulti]()},
		{"direct Region multi required", reflect.TypeFor[directRegionMultiRequired]()},
		{"slice Region single", reflect.TypeFor[sliceRegionSingle]()},
		{"slice Region single required", reflect.TypeFor[sliceRegionSingleRequired]()},
		{"direct Landmark multi", reflect.TypeFor[directLandmarkMulti]()},
		{"direct Landmark multi required", reflect.TypeFor[directLandmarkMultiRequired]()},
		{"slice Landmark single", reflect.TypeFor[sliceLandmarkSingle]()},
		{"slice Landmark single required", reflect.TypeFor[sliceLandmarkSingleRequired]()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			subject, check := capture.TypeReflection(tc.typ).AsStruct()
			if !check.OK() {
				t.Fatalf("capture subject: %s", check.Why())
			}
			_, validation := lane.Compile(subject, nil)
			if validation.IsOkay() {
				t.Fatal("compilation succeeded for representation/shape mismatch")
			}
			if !strings.Contains(validation.Trace(), string(SpotKindMismatchCode)) {
				t.Fatalf("failure lacks %q diagnostic: %s", SpotKindMismatchCode, validation.Trace())
			}
		})
	}
}

type duplicateOptionalSpot struct {
	legend.Lane
	First  occurrenceLandmark `sail:"spot" spot:"single"`
	Second occurrenceLandmark `sail:"spot" spot:"single"`
}
type duplicateRequiredSpot struct {
	legend.Lane
	First  occurrenceLandmark `sail:"spot" spot:"single"`
	Second occurrenceLandmark `sail:"spot" spot:"single,required"`
}

func TestSpotRequiredDoesNotChangeSelectorUniqueness(t *testing.T) {
	lane, err := define.Struct[LaneDefinition, declarations.Lane](nil)
	if err != nil {
		t.Fatalf("define lane: %v", err)
	}
	for _, typ := range []reflect.Type{reflect.TypeFor[duplicateOptionalSpot](), reflect.TypeFor[duplicateRequiredSpot]()} {
		subject, check := capture.TypeReflection(typ).AsStruct()
		if !check.OK() {
			t.Fatalf("capture %s: %s", typ, check.Why())
		}
		_, validation := lane.Compile(subject, nil)
		if validation.IsOkay() || !strings.Contains(validation.Trace(), "duplicate_spot_selector") || !strings.Contains(validation.Trace(), `fields "First" and "Second"`) {
			t.Fatalf("expected duplicate selector for %s, got: %s", typ, validation.Trace())
		}
	}
}
