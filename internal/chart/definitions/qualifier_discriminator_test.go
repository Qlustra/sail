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

// The graph-level validator must inspect all reachable Region and Landmark
// bodies, not just the definition currently being compiled.
type qdCollisionRegion struct {
	legend.Region `sail:"region" family:"f" discriminate:"shared,other"`
	Presence      testBody `sail:"qualifier" symbol:"only-here"`
}

type collisionLandmark struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:"unused"`
	Presence        testBody `sail:"qualifier" symbol:"shared"`
}

type sameBodyCollision struct {
	legend.Region `sail:"region" family:"f" discriminate:"same"`
	Presence      testBody `sail:"qualifier" symbol:"same"`
}

type sameBodyCollisionGraph struct {
	legend.Legend `sail:"legend" mark:"x"`
	Lane          sameBodyCollisionLane `sail:"lane"`
}

type sameBodyCollisionLane struct {
	legend.Lane
	Region sameBodyCollision `sail:"spot" spot:"single"`
}

type nestedCollisionRegion struct {
	legend.Region `sail:"region" family:"nested" discriminate:"deep"`
	Child         nestedCollisionLane `sail:"lane"`
}

type nestedCollisionLane struct {
	legend.Lane
	Leaf nestedCollisionLandmark `sail:"spot" spot:"single"`
}

type nestedCollisionLandmark struct {
	legend.Landmark `sail:"landmark" family:"nested" discriminate:"other"`
	Deep            testBody `sail:"qualifier" symbol:"deep"`
}

type nestedCollisionGraph struct {
	legend.Legend `sail:"legend" mark:"x"`
	Lane          nestedCollisionGraphLane `sail:"lane"`
}

type nestedCollisionGraphLane struct {
	legend.Lane
	Root nestedCollisionRegion `sail:"spot" spot:"single"`
}

type differentFamilyGraph struct {
	legend.Legend `sail:"legend" mark:"x"`
	Lane          differentFamilyLane `sail:"lane"`
}

type differentFamilyLane struct {
	legend.Lane
	First  qdCollisionRegion   `sail:"spot" spot:"single"`
	Second otherFamilyLandmark `sail:"spot" spot:"single"`
}

type otherFamilyLandmark struct {
	legend.Landmark `sail:"landmark" family:"g" discriminate:"unused"`
	Presence        legend.Qualifier `sail:"qualifier" symbol:"shared"`
}

type collisionLane struct {
	legend.Lane
	Region   qdCollisionRegion `sail:"spot" spot:"single"`
	Landmark collisionLandmark `sail:"spot" spot:"single"`
}

type collisionGraphRegionFirst struct {
	legend.Legend `sail:"legend" mark:"x"`
	Lane          collisionLane `sail:"lane"`
}

type collisionGraphLandmarkFirst struct {
	legend.Legend `sail:"legend" mark:"x"`
	Lane          reverseCollisionLane `sail:"lane"`
}

type reverseCollisionLane struct {
	legend.Lane
	Landmark collisionLandmark `sail:"spot" spot:"single"`
	Region   qdCollisionRegion `sail:"spot" spot:"single"`
}

type repeatedVocabularyRegion struct {
	legend.Region `sail:"region" family:"f" discriminate:"one,two"`
	Feature       legend.Feature   `sail:"feature" symbol:"one"`
	One           legend.Qualifier `sail:"qualifier" symbol:"choice"`
	Two           legend.Qualifier `sail:"qualifier" symbol:"value"`
}

type repeatedVocabularyGraph struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         repeatedVocabularyLane `sail:"lane"`
	Second        secondRepeatedLane     `sail:"lane"`
}

type secondRepeatedRegion struct {
	legend.Region `sail:"region" family:"f" discriminate:"one,four"`
	One           legend.Qualifier `sail:"qualifier" symbol:"choice"`
	Two           legend.Qualifier `sail:"qualifier" symbol:"value"`
}

type secondRepeatedLane struct {
	legend.Lane
	First secondRepeatedRegion `sail:"spot" spot:"single"`
}

type repeatedVocabularyLane struct {
	legend.Lane
	First repeatedVocabularyRegion `sail:"spot" spot:"single"`
}

type sharedRecursiveRegion struct {
	legend.Region `sail:"region" family:"f" discriminate:"shared"`
	Presence      testBody            `sail:"qualifier" symbol:"not-shared"`
	Children      sharedRecursiveLane `sail:"lane"`
}

type sharedRecursiveLane struct {
	legend.Lane
	Back []sharedRecursiveRegion `sail:"spot" spot:"multi"`
	Leaf sharedQualifierLandmark `sail:"spot" spot:"single"`
}

type sharedRecursiveGraph struct {
	legend.Legend `sail:"legend" mark:"x"`
	Lane          sharedRecursiveLane `sail:"lane"`
}

type sharedQualifierLandmark struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:"other"`
	Presence        testBody `sail:"qualifier" symbol:"different"`
}

func compileDiscriminatorTestGraph(t *testing.T, typ reflect.Type) (*declarations.Legend, bool) {
	t.Helper()
	definition, err := define.Struct[LegendDefinition, declarations.Legend](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.TypeReflection(typ).AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := definition.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatalf("fixture compilation failed: %s", validation.Trace())
	}
	value, check := ctx.Access[declarations.Legend](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if check := declarations.ValidateLaneExploration(value); !check.OK() {
		t.Fatalf("fixture failed lane exploration validation: %s", check.Why())
	}
	return value, true
}

func TestValidateQualifierDiscriminatorsGraph(t *testing.T) {
	for _, tc := range []struct {
		typ  reflect.Type
		want []string
	}{
		{reflect.TypeFor[sameBodyCollisionGraph](), []string{`family="f"`, `token="same"`, `"Presence"`, `use "Region"`}},
		{reflect.TypeFor[collisionGraphRegionFirst](), []string{`family="f"`, `token="shared"`, `"Presence"`, `use "Region"`}},
		{reflect.TypeFor[collisionGraphLandmarkFirst](), []string{`family="f"`, `token="shared"`, `"Presence"`, `use "Region"`}},
		{reflect.TypeFor[nestedCollisionGraph](), []string{`family="nested"`, `token="deep"`, `"Deep"`, `use "Root"`}},
	} {
		t.Run(tc.typ.Name(), func(t *testing.T) {
			graph, compiled := compileDiscriminatorTestGraph(t, tc.typ)
			if !compiled {
				t.Fatalf("fixture compilation failed before graph validation")
			}
			check := declarations.ValidateQualifierDiscriminators(graph)
			if check.OK() {
				t.Fatal("expected discriminator/qualifier collision")
			}
			why := check.Why()
			for _, fragment := range append(tc.want, "qualifier field", "discriminator in") {
				if !strings.Contains(why, fragment) {
					t.Fatalf("diagnostic %q missing %q", why, fragment)
				}
			}
		})
	}
}

func TestValidateQualifierDiscriminatorsAllowsDistinctFamilyAndRepeatedTokens(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[differentFamilyGraph](), reflect.TypeFor[repeatedVocabularyGraph](), reflect.TypeFor[sharedRecursiveGraph]()} {
		t.Run(typ.Name(), func(t *testing.T) {
			graph, compiled := compileDiscriminatorTestGraph(t, typ)
			if !compiled {
				t.Fatal("unreachable")
			}
			if check := declarations.ValidateQualifierDiscriminators(graph); !check.OK() {
				t.Fatalf("unexpected graph collision: %s", check.Why())
			}
		})
	}
}

func TestQualifierDiscriminatorGraphCompilationDeterministic(t *testing.T) {
	first, ok := compileDiscriminatorTestGraph(t, reflect.TypeFor[collisionGraphRegionFirst]())
	if !ok {
		t.Fatal("first fixture did not compile")
	}
	second, ok := compileDiscriminatorTestGraph(t, reflect.TypeFor[collisionGraphLandmarkFirst]())
	if !ok {
		t.Fatal("second fixture did not compile")
	}
	firstWhy := declarations.ValidateQualifierDiscriminators(first).Why()
	secondWhy := declarations.ValidateQualifierDiscriminators(second).Why()
	for i := 0; i < 10; i++ {
		if got := declarations.ValidateQualifierDiscriminators(first).Why(); got != firstWhy {
			t.Fatalf("diagnostic changed across validations: %q != %q", got, firstWhy)
		}
		if got := declarations.ValidateQualifierDiscriminators(second).Why(); got != secondWhy {
			t.Fatalf("reverse diagnostic changed across validations: %q != %q", got, secondWhy)
		}
	}
	if firstWhy != secondWhy {
		t.Fatalf("diagnostic depends on encounter order:\nfirst:  %s\nsecond: %s", firstWhy, secondWhy)
	}
}
