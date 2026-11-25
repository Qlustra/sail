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

type testBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type noDefaultBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Other() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type wrongResultBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[int, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type wrongArityBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default(int) string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type wrongOutputsBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default() (string, error) }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type panicProvider struct{}

func (panicProvider) Default() string { panic("provider must not be invoked during compilation") }

type panicProviderBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type wrongResultRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             wrongResultBody `sail:"qualifier" symbol:"q"`
}

type wrongArityRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             wrongArityBody `sail:"qualifier" symbol:"q"`
}

type wrongOutputsRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             wrongOutputsBody `sail:"qualifier" symbol:"q"`
}

type panicProviderRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             panicProviderBody `sail:"qualifier" symbol:"q"`
}

type missingBodyTag struct {
	legend.Qualifier
	Name legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type missingBodyTagRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             missingBodyTag `sail:"qualifier" symbol:"q"`
}

type missingMarkerBody struct {
	Name legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type missingMarkerRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             missingMarkerBody `sail:"qualifier" symbol:"q"`
}

type emptyBody struct {
	legend.Qualifier `sail:"qualifier"`
}

type bareRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             legend.Qualifier `sail:"qualifier" symbol:"q"`
}

type barrelQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             testBody `sail:"qualifier" symbol:"q"`
}

type optionalBarrelBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default"`
	Required         legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type optionalBarrelRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             optionalBarrelBody `sail:"qualifier" symbol:"optional"`
}

// A shared body describes Barrels; the symbol is attached to each use.
type sharedQualifierBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default() string }] `sail:"barrel" populate:"Default" barrel:"required"`
}

type sharedQualifierRegion struct {
	legend.Region `sail:"region" family:"test"`
	First         sharedQualifierBody `sail:"qualifier" symbol:"first"`
	Second        sharedQualifierBody `sail:"qualifier" symbol:"second"`
}

type missingMethodRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             noDefaultBody `sail:"qualifier" symbol:"q"`
}

type missingPopulateBody struct {
	legend.Qualifier `sail:"qualifier"`
	Name             legend.Barrel[string, interface{ Default() string }] `sail:"barrel"`
}

type missingPopulateRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             missingPopulateBody `sail:"qualifier" symbol:"q"`
}

type qualifierEmptyRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             emptyBody `sail:"qualifier" symbol:"q"`
}

type wrongMarkerRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             legend.Feature `sail:"qualifier" symbol:"q"`
}

type duplicateBody struct {
	legend.Qualifier `sail:"qualifier"`
	Other            legend.Qualifier `sail:"qualifier"`
}

type duplicateRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             duplicateBody `sail:"qualifier" symbol:"q"`
}

type collisionRegion struct {
	legend.Region `sail:"region" family:"test"`
	Q             legend.Qualifier `sail:"qualifier" symbol:"same"`
	F             legend.Feature   `sail:"feature" symbol:"same"`
}

type landmarkRegion struct {
	legend.Landmark `sail:"landmark" family:"test"`
	Q               legend.Qualifier `sail:"qualifier" symbol:"q"`
}

func TestQualifierCompilation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typ      any
		okay     bool
		landmark bool
	}{
		{"bare qualifier", bareRegion{}, true, false},
		{"barrel populate", barrelQualifierRegion{}, true, false},
		{"optional barrel", optionalBarrelRegion{}, true, false},
		{"barrel invocation is deferred", panicProviderRegion{}, true, false},
		{"wrong result", wrongResultRegion{}, false, false},
		{"wrong arity", wrongArityRegion{}, false, false},
		{"wrong output count", wrongOutputsRegion{}, false, false},
		{"missing body marker", missingMarkerRegion{}, false, false},
		{"missing body tag", missingBodyTagRegion{}, false, false},
		{"missing method", missingMethodRegion{}, false, false},
		{"missing populate", missingPopulateRegion{}, false, false},
		{"empty body", qualifierEmptyRegion{}, false, false},
		{"wrong marker", wrongMarkerRegion{}, false, false},
		{"duplicate symbol", duplicateRegion{}, false, false},
		{"same-symbol feature", collisionRegion{}, true, false},
		{"landmark", landmarkRegion{}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var okay bool
			if tc.landmark {
				definition, err := define.Struct[LandmarkDefinition, declarations.Landmark](nil)
				if err != nil {
					t.Fatal(err)
				}
				subject, check := capture.TypeReflection(reflect.TypeOf(tc.typ)).AsStruct()
				if !check.OK() {
					t.Fatal(check.Why())
				}
				_, validation := definition.Compile(subject, nil)
				okay = validation.IsOkay()
			} else {
				definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
				if err != nil {
					t.Fatal(err)
				}
				subject, check := capture.TypeReflection(reflect.TypeOf(tc.typ)).AsStruct()
				if !check.OK() {
					t.Fatal(check.Why())
				}
				_, validation := definition.Compile(subject, nil)
				okay = validation.IsOkay()
			}
			if okay != tc.okay {
				t.Fatalf("validation okay = %v; want %v", okay, tc.okay)
			}
		})
	}
}

func TestQualifierDuplicateSymbolsAreLocalToHost(t *testing.T) {
	type duplicate struct {
		legend.Region `sail:"region" family:"test"`
		First         legend.Qualifier `sail:"qualifier" symbol:"repeat"`
		Second        legend.Qualifier `sail:"qualifier" symbol:"repeat"`
	}
	d, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[duplicate]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	_, validation := d.Compile(subject, nil)
	if validation.IsOkay() || !strings.Contains(validation.Trace(), "duplicate_qualifier_symbol") || !strings.Contains(validation.Trace(), `fields "First" and "Second"`) {
		t.Fatalf("expected local duplicate symbol diagnostic, got: %s", validation.Trace())
	}
}

func TestQualifierBarrelAvailability(t *testing.T) {
	d, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[optionalBarrelRegion]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := d.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatal(validation.Trace())
	}
	body, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	barrels := body.Qualifiers[0].Declaration.Barrels
	if len(barrels) != 2 || barrels[0].Required || !barrels[1].Required || barrels[0] == barrels[1] {
		t.Fatalf("incorrect qualifier barrel availability: %s", body.Trace())
	}
	trace := body.Trace()
	if !strings.Contains(trace, "barrel required=false") || !strings.Contains(trace, "barrel required=true") {
		t.Fatalf("trace missing availability: %s", trace)
	}
}

func TestQualifierSharedBodyHasPerUseSymbols(t *testing.T) {
	definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[sharedQualifierRegion]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := definition.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatalf("compile shared qualifier body: %s", validation.Trace())
	}
	body, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(body.Qualifiers) != 2 {
		t.Fatalf("qualifier uses = %d; want 2", len(body.Qualifiers))
	}
	first, second := body.Qualifiers[0], body.Qualifiers[1]
	if first.Symbol != "first" || second.Symbol != "second" {
		t.Fatalf("qualifier symbols = %q, %q", first.Symbol, second.Symbol)
	}
	if first.Declaration != second.Declaration {
		t.Fatal("shared body uses do not retain the same qualifier declaration")
	}
	if first.Property == second.Property {
		t.Fatal("shared body uses unexpectedly retain the same use property")
	}
	trace := body.Trace()
	if !strings.Contains(trace, "first") || !strings.Contains(trace, "second") {
		t.Fatalf("trace does not retain both use symbols: %s", trace)
	}
	if !strings.Contains(trace, "(shared)") {
		t.Fatalf("trace does not identify the shared qualifier body: %s", trace)
	}
}
