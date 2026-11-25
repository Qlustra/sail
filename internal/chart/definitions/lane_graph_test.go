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

type graphLandmarkA struct {
	legend.Landmark `sail:"landmark" family:"a"`
}
type graphLandmarkB struct {
	legend.Landmark `sail:"landmark" family:"b"`
}
type graphLandmarkADiscriminateA struct {
	legend.Landmark `sail:"landmark" family:"a" discriminate:"a"`
}
type graphLandmarkADiscriminateAB struct {
	legend.Landmark `sail:"landmark" family:"a" discriminate:"a,b"`
}

type graphLaneMatch struct {
	legend.Lane
	Match graphLandmarkA `sail:"spot" spot:"single"`
}
type graphLaneDifferent struct {
	legend.Lane
	Other graphLandmarkB `sail:"spot" spot:"single"`
}
type graphLaneRequiredDifferent struct {
	legend.Lane
	Other graphLandmarkB `sail:"spot" spot:"single,required"`
}
type graphLaneRequiredMatch struct {
	legend.Lane
	Match graphLandmarkA `sail:"spot" spot:"single,required"`
}
type graphLaneRequiredAfterOptional struct {
	legend.Lane
	Optional graphLandmarkB                `sail:"spot" spot:"single"`
	Required []graphLandmarkADiscriminateA `sail:"spot" spot:"multi,required"`
}
type graphLaneDiscriminateA struct {
	legend.Lane
	Match graphLandmarkADiscriminateA `sail:"spot" spot:"single"`
}
type graphLaneDiscriminateAB struct {
	legend.Lane
	Match graphLandmarkADiscriminateAB `sail:"spot" spot:"single"`
}

type graphSameDefault struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch `sail:"lane"`
	Second        graphLaneMatch `sail:"lane"`
}
type graphDefaultEager struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch `sail:"lane"`
	Second        graphLaneMatch `sail:"lane" exploration:"eager"`
}
type graphEagerDefault struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch `sail:"lane" exploration:"eager"`
	Second        graphLaneMatch `sail:"lane"`
}
type graphOptionalBetween struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch     `sail:"lane"`
	Between       graphLaneDifferent `sail:"lane"`
	Last          graphLaneMatch     `sail:"lane"`
}
type graphRequiredBetween struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch             `sail:"lane"`
	Between       graphLaneRequiredDifferent `sail:"lane"`
	Last          graphLaneMatch             `sail:"lane"`
}
type graphRequiredBoundary struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch         `sail:"lane"`
	Boundary      graphLaneRequiredMatch `sail:"lane"`
	Last          graphLaneMatch         `sail:"lane"`
}
type graphRequiredCurrent struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneRequiredMatch `sail:"lane"`
	Last          graphLaneMatch         `sail:"lane"`
}
type graphMixedRequiredBarrier struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch                 `sail:"lane"`
	Between       graphLaneRequiredAfterOptional `sail:"lane"`
	Last          graphLaneMatch                 `sail:"lane"`
}
type graphRegionA struct {
	legend.Region `sail:"region" family:"a"`
}
type graphRegionLane struct {
	legend.Lane
	Region graphRegionA `sail:"spot" spot:"single"`
}
type graphCrossKindOverlap struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch  `sail:"lane"`
	Second        graphRegionLane `sail:"lane"`
}
type graphDifferentDiscriminators struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneDiscriminateA  `sail:"lane"`
	Second        graphLaneDiscriminateAB `sail:"lane"`
}
type graphSharedDefaultEager struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch `sail:"lane"`
	Second        graphLaneMatch `sail:"lane" exploration:"eager"`
}
type graphSharedEagerDefault struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneMatch `sail:"lane" exploration:"eager"`
	Second        graphLaneMatch `sail:"lane"`
}

type graphRecursiveRegion struct {
	legend.Region `sail:"region" family:"node"`
	Children      graphRecursiveLane `sail:"lane"`
}
type graphRecursiveLane struct {
	legend.Lane
	Children []graphRecursiveRegion `sail:"spot" spot:"multi"`
}
type graphNestedGood struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphRecursiveLane `sail:"lane" exploration:"eager"`
	Second        graphRecursiveLane `sail:"lane"`
}
type graphBadRegion struct {
	legend.Region `sail:"region" family:"bad"`
	First         graphLaneMatch        `sail:"lane"`
	Second        graphLaneMatch        `sail:"lane" exploration:"eager"`
	Recursive     graphBadRecursiveLane `sail:"lane"`
}
type graphBadRecursiveLane struct {
	legend.Lane
	Children []graphBadRegion `sail:"spot" spot:"multi"`
}
type graphNestedBad struct {
	legend.Legend `sail:"legend" mark:"x"`
	Outer         graphBadRecursiveLane `sail:"lane"`
}
type graphLandmarkReverse struct {
	legend.Landmark `sail:"landmark" family:"a" discriminate:"b,a"`
}
type graphOtherSpotLane struct {
	legend.Lane
	OtherSpot graphLandmarkReverse `sail:"spot" spot:"single"`
}
type graphNormalizedCollision struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         graphLaneDiscriminateAB `sail:"lane"`
	Second        graphOtherSpotLane      `sail:"lane"`
}

func TestValidateLaneExplorationCompiledLegendGraphs(t *testing.T) {
	cases := []struct {
		name       string
		typ        reflect.Type
		pass       bool
		diagnostic []string
	}{
		{"identical default/default", reflect.TypeFor[graphSameDefault](), false, []string{"First", "Second", "Match", `family="a"`}},
		{"default/eager", reflect.TypeFor[graphDefaultEager](), false, []string{"First", "Second", "Match", `family="a"`}},
		{"eager/default", reflect.TypeFor[graphEagerDefault](), true, nil},
		{"optional nonmatching intervening", reflect.TypeFor[graphOptionalBetween](), false, []string{"First", "Last", "Match", `family="a"`}},
		{"required nonmatching intervening", reflect.TypeFor[graphRequiredBetween](), true, nil},
		{"required matching boundary", reflect.TypeFor[graphRequiredBoundary](), false, []string{"First", "Boundary", "Match", `family="a"`}},
		{"required current single then matching", reflect.TypeFor[graphRequiredCurrent](), false, []string{"First", "Last", "Match", `family="a"`}},
		{"required multi after optional Spot forms a barrier", reflect.TypeFor[graphMixedRequiredBarrier](), true, nil},
		{"Landmark versus Region selector overlap", reflect.TypeFor[graphCrossKindOverlap](), false, []string{"First", "Second", "Match", "Region", `family="a"`}},
		{"discriminator a versus a,b", reflect.TypeFor[graphDifferentDiscriminators](), true, nil},
		{"shared lane default/eager", reflect.TypeFor[graphSharedDefaultEager](), false, []string{"First", "Second", "Match", `family="a"`}},
		{"shared lane eager/default", reflect.TypeFor[graphSharedEagerDefault](), true, nil},
		{"shared recursive region", reflect.TypeFor[graphNestedGood](), true, nil},
		{"nested recursive region conflict", reflect.TypeFor[graphNestedBad](), false, []string{"First", "Second", "Match", `family="a"`}},
		{"normalized selector and distinct spot fields", reflect.TypeFor[graphNormalizedCollision](), false, []string{"First", "Second", "Match", "OtherSpot", `family="a"`, `discriminate="a,b"`}},
	}
	d, err := define.Struct[LegendDefinition, declarations.Legend](nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subject, check := capture.TypeReflection(tc.typ).AsStruct()
			if !check.OK() {
				t.Fatal(check.Why())
			}
			ctx, compiled := d.Compile(subject, nil)
			if !compiled.IsOkay() {
				t.Fatalf("compile: %s", compiled.Trace())
			}
			body, check := ctx.Access[declarations.Legend](define.DeclarationKey)
			if !check.OK() {
				t.Fatal(check.Why())
			}
			validation := declarations.ValidateLaneExploration(body)
			if validation.OK() != tc.pass {
				t.Fatalf("unexpected validation result: %s", validation.Why())
			}
			if !tc.pass {
				for i := 0; i < 10; i++ {
					if got := declarations.ValidateLaneExploration(body).Why(); got != validation.Why() {
						t.Fatalf("nondeterministic diagnostic: %q vs %q", got, validation.Why())
					}
				}
				for _, text := range tc.diagnostic {
					if !strings.Contains(validation.Why(), text) {
						t.Errorf("diagnostic %q missing %q", validation.Why(), text)
					}
				}
			}
		})
	}
}
