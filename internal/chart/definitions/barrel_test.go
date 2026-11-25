package definitions

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/legend"
)

type barrelCapability interface {
	Read() int
}
type barrelInput interface{ Read(int) int }
type barrelNoResult interface{ Read() }
type barrelManyResults interface{ Read() (int, error) }
type barrelWrongResult interface{ Read() string }
type barrelInvalidCapability struct{}
type barrelNoMethod interface{ Missing() int }
type barrelUnexported interface{ hidden() int }
type barrelNamedEmpty interface{}
type barrelDefaultable[T any] interface{ Default() T }

type barrelUnicodeCapability interface{ Étiquette() int }

type barrelAnyDynamic struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[any, any] `sail:"barrel" populate:"Default"`
}

type barrelNamedEmptyDynamic struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[string, barrelNamedEmpty] `sail:"barrel" populate:"Default"`
}

type barrelDefaultableAny struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[any, barrelDefaultable[any]] `sail:"barrel" populate:"Default"`
}

type barrelEmptyDynamicSelector struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, any] `sail:"barrel" populate:""`
}

type barrelMalformedDynamicSelector struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, any] `sail:"barrel" populate:"Default!"`
}

type barrelUnexportedDynamicSelector struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, any] `sail:"barrel" populate:"default"`
}

type barrelUnicodeDynamicSelector struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, any] `sail:"barrel" populate:"Étiquette"`
}

type barrelUnicodeSelector struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelUnicodeCapability] `sail:"barrel" populate:"Étiquette"`
}
type barrelSource struct {
	legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read"`
}

type regionBarrelRequired struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:"required"`
}
type landmarkBarrelOptional struct {
	legend.Landmark `sail:"landmark" family:"l"`
	Value           legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:"optional"`
}
type regionBarrelMissing struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read"`
}
type regionBarrelEmpty struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:""`
}
type regionBarrelBadTag struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:"Required"`
}
type regionBarrelUnknown struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:"unknown"`
}
type regionBarrelLeadingSpace struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:" required"`
}
type regionBarrelTrailingSpace struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:"required "`
}
type regionBarrelCompound struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Read" barrel:"optional,required"`
}
type regionBarrelBadMethod struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:"Missing" barrel:"required"`
}
type regionBarrelMissingPopulate struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" barrel:"required"`
}
type regionBarrelEmptyPopulate struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"barrel" populate:""`
}
type regionBarrelBadInput struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelInput] `sail:"barrel" populate:"Read"`
}
type regionBarrelNoResult struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelNoResult] `sail:"barrel" populate:"Read"`
}
type regionBarrelManyResults struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelManyResults] `sail:"barrel" populate:"Read"`
}
type regionBarrelWrongResult struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelWrongResult] `sail:"barrel" populate:"Read"`
}
type regionBarrelAnyResult struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[any, barrelCapability] `sail:"barrel" populate:"Read"`
}
type regionBarrelInvalidCapability struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelInvalidCapability] `sail:"barrel" populate:"Read"`
}
type regionBarrelNoMethod struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelNoMethod] `sail:"barrel" populate:"Read"`
}
type regionBarrelUnexported struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelUnexported] `sail:"barrel" populate:"hidden"`
}
type regionBarrelRepeated struct {
	legend.Region `sail:"region" family:"r"`
	First         barrelSource `sail:"barrel" populate:"Read"`
	Second        barrelSource `sail:"barrel" populate:"Read"`
}
type barrelWrongType struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[int, barrelCapability] `sail:"feature" populate:"Read"`
}
type barrelTagWrongType struct {
	legend.Region `sail:"region" family:"r"`
	Value         string `sail:"barrel" barrel:"required" populate:"Read"`
}

func compileBarrelFixture(t *testing.T, typ reflect.Type, region bool) (*declarations.Region, *declarations.Landmark, string, bool) {
	t.Helper()
	if region {
		definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
		if err != nil {
			t.Fatal(err)
		}
		subject, check := capture.TypeReflection(typ).AsStruct()
		if !check.OK() {
			t.Fatal(check.Why())
		}
		compiled, validation := definition.Compile(subject, nil)
		if validation.IsOkay() {
			value, access := compiled.Access[declarations.Region](define.DeclarationKey)
			if !access.OK() {
				t.Fatal(access.Why())
			}
			return value, nil, value.Trace(), true
		}
		return nil, nil, validation.Trace(), false
	}
	definition, err := define.Struct[LandmarkDefinition, declarations.Landmark](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.TypeReflection(typ).AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	compiled, validation := definition.Compile(subject, nil)
	if validation.IsOkay() {
		value, access := compiled.Access[declarations.Landmark](define.DeclarationKey)
		if !access.OK() {
			t.Fatal(access.Why())
		}
		return nil, value, value.Trace(), true
	}
	return nil, nil, validation.Trace(), false
}

func TestBarrelCompilationMetadata(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		typ                     reflect.Type
		region, required, valid bool
	}{
		{"Region required", reflect.TypeFor[regionBarrelRequired](), true, true, true},
		{"Landmark optional", reflect.TypeFor[landmarkBarrelOptional](), false, false, true},
		{"missing is optional", reflect.TypeFor[regionBarrelMissing](), true, false, true},
		{"empty tag rejected", reflect.TypeFor[regionBarrelEmpty](), true, false, false},
		{"invalid vocabulary", reflect.TypeFor[regionBarrelBadTag](), true, false, false},
		{"invalid method", reflect.TypeFor[regionBarrelBadMethod](), true, false, false},
		{"missing populate", reflect.TypeFor[regionBarrelMissingPopulate](), true, false, false},
		{"empty populate", reflect.TypeFor[regionBarrelEmptyPopulate](), true, false, false},
		{"invalid I", reflect.TypeFor[regionBarrelInvalidCapability](), true, false, false},
		{"missing method", reflect.TypeFor[regionBarrelNoMethod](), true, false, false},
		{"unexported method", reflect.TypeFor[regionBarrelUnexported](), true, false, false},
		{"input arity", reflect.TypeFor[regionBarrelBadInput](), true, false, false},
		{"zero outputs", reflect.TypeFor[regionBarrelNoResult](), true, false, false},
		{"multiple outputs", reflect.TypeFor[regionBarrelManyResults](), true, false, false},
		{"nonassignable result", reflect.TypeFor[regionBarrelWrongResult](), true, false, false},
		{"any accepts result", reflect.TypeFor[regionBarrelAnyResult](), true, false, true},
		{"wrong tag", reflect.TypeFor[barrelWrongType](), true, false, false},
		{"wrong type with barrel tag", reflect.TypeFor[barrelTagWrongType](), true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			region, landmark, trace, success := compileBarrelFixture(t, tc.typ, tc.region)
			if success != tc.valid {
				t.Fatalf("success=%t trace=%s", success, trace)
			}
			if !success {
				if !strings.Contains(strings.ToLower(trace), "barrel") && tc.name != "wrong tag" {
					t.Fatalf("trace omits barrel context: %s", trace)
				}
				return
			}
			var barrels int
			if region != nil {
				barrels = len(region.Barrels)
			} else {
				barrels = len(landmark.Barrels)
			}
			if barrels != 1 {
				t.Fatalf("barrels=%d", barrels)
			}
			if !strings.Contains(trace, "barrel required="+strconv.FormatBool(tc.required)) {
				t.Fatalf("trace lacks required state: %s", trace)
			}
			for _, part := range []string{"(field Value)", `populate="Read"`, "capability=definitions.barrelCapability", "result="} {
				if !strings.Contains(trace, part) {
					t.Fatalf("trace lacks %q: %s", part, trace)
				}
			}
		})
	}
}

func TestBarrelDynamicAndLiteralInterfaceBindings(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  reflect.Type
		ok   bool
		want string
	}{
		{"any dynamic binding", reflect.TypeFor[barrelAnyDynamic](), true, "capability=interface {}"},
		{"named empty dynamic binding", reflect.TypeFor[barrelNamedEmptyDynamic](), true, "capability=definitions.barrelNamedEmpty"},
		{"literal Defaultable any", reflect.TypeFor[barrelDefaultableAny](), true, "capability=definitions.barrelDefaultable[interface {}]"},
		{"empty dynamic selector", reflect.TypeFor[barrelEmptyDynamicSelector](), false, "nonempty populate tag"},
		{"malformed dynamic selector", reflect.TypeFor[barrelMalformedDynamicSelector](), false, "exported Go identifier"},
		{"unexported dynamic selector", reflect.TypeFor[barrelUnexportedDynamicSelector](), false, "exported Go identifier"},
		{"Unicode dynamic selector", reflect.TypeFor[barrelUnicodeDynamicSelector](), true, "populate=\"Étiquette\""},
		{"exported Unicode interface method", reflect.TypeFor[barrelUnicodeSelector](), true, "populate=\"Étiquette\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, trace, success := compileBarrelFixture(t, tc.typ, true)
			if success != tc.ok {
				t.Fatalf("success=%t trace=%s", success, trace)
			}
			if !strings.Contains(trace, tc.want) {
				t.Fatalf("trace does not contain %q: %s", tc.want, trace)
			}
		})
	}
	for _, tc := range []struct {
		name string
		typ  reflect.Type
		want reflect.Type
	}{
		{"any", reflect.TypeFor[barrelAnyDynamic](), reflect.TypeFor[any]()},
		{"named empty", reflect.TypeFor[barrelNamedEmptyDynamic](), reflect.TypeFor[string]()},
	} {
		t.Run("retained result and nil Method/"+tc.name, func(t *testing.T) {
			region, _, trace, success := compileBarrelFixture(t, tc.typ, true)
			if !success {
				t.Fatalf("compile: %s", trace)
			}
			if len(region.Barrels) != 1 {
				t.Fatalf("barrels=%d", len(region.Barrels))
			}
			binding := region.Barrels[0].ValueBinding
			if binding.Shape.Result.CapturedType() != tc.want {
				t.Fatalf("retained result=%v; want %v", binding.Shape.Result.CapturedType(), tc.want)
			}
			if binding.Method != nil {
				t.Fatalf("dynamic empty-interface binding has Method=%v; want nil", binding.Method)
			}
		})
	}
	region, _, trace, success := compileBarrelFixture(t, reflect.TypeFor[barrelDefaultableAny](), true)
	if !success {
		t.Fatalf("compile literal Defaultable[any]: %s", trace)
	}
	if len(region.Barrels) != 1 || region.Barrels[0].Method == nil {
		t.Fatalf("literal interface method was not retained: %s", trace)
	}
	if region.Barrels[0].Method.ResultAssignableTo(reflect.TypeFor[int]()) {
		t.Fatal("literal Default() any must not be treated as assignable to int")
	}
}

func TestBarrelStrictVocabulary(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[regionBarrelEmpty](), reflect.TypeFor[regionBarrelBadTag](), reflect.TypeFor[regionBarrelUnknown](), reflect.TypeFor[regionBarrelLeadingSpace](), reflect.TypeFor[regionBarrelTrailingSpace](), reflect.TypeFor[regionBarrelCompound]()} {
		t.Run(typ.Name(), func(t *testing.T) {
			_, _, trace, success := compileBarrelFixture(t, typ, true)
			if success {
				t.Fatalf("accepted %s: %s", typ, trace)
			}
		})
	}
}

func TestBarrelRepeatedSourceRetainsDistinctDestinations(t *testing.T) {
	region, _, trace, success := compileBarrelFixture(t, reflect.TypeFor[regionBarrelRepeated](), true)
	if !success {
		t.Fatalf("compile repeated barrels: %s", trace)
	}
	if len(region.Barrels) != 2 {
		t.Fatalf("got %d barrels, want 2: %s", len(region.Barrels), trace)
	}
	if region.Barrels[0] == region.Barrels[1] {
		t.Fatal("repeated source fields share a destination carrier")
	}
	if region.Barrels[0].Encounter.Property.FieldName() != "First" || region.Barrels[1].Encounter.Property.FieldName() != "Second" {
		t.Fatalf("barrel destinations: %s, %s", region.Barrels[0].Encounter.Property.FieldName(), region.Barrels[1].Encounter.Property.FieldName())
	}
}

func TestBarrelCompilationDoesNotInvokeProvider(t *testing.T) {
	barrelProviderCalls = 0
	for _, typ := range []reflect.Type{reflect.TypeFor[regionBarrelRequired](), reflect.TypeFor[regionBarrelAnyResult]()} {
		_, _, trace, success := compileBarrelFixture(t, typ, true)
		if !success {
			t.Fatalf("compile %s: %s", typ, trace)
		}
	}
	if barrelProviderCalls != 0 {
		t.Fatalf("provider invoked %d times", barrelProviderCalls)
	}
}

var barrelProviderCalls int

type barrelProvider struct{}

func (barrelProvider) Read() int { barrelProviderCalls++; return 0 }

var _ barrelCapability = barrelProvider{}

type barrelSelfReturningProvider struct{}

var barrelSelfReturningProviderCalls int

func (barrelSelfReturningProvider) Default() barrelSelfReturningProvider {
	barrelSelfReturningProviderCalls++
	return barrelSelfReturningProvider{}
}

type barrelSelfReturningFixture struct {
	legend.Region `sail:"region" family:"r"`
	Value         legend.Barrel[any, any] `sail:"barrel" populate:"Default"`
}

func TestBarrelCompilationDoesNotInvokeSelfReturningProvider(t *testing.T) {
	barrelSelfReturningProviderCalls = 0
	_, _, trace, success := compileBarrelFixture(t, reflect.TypeFor[barrelSelfReturningFixture](), true)
	if !success {
		t.Fatalf("compile self-returning provider: %s", trace)
	}
	if barrelSelfReturningProviderCalls != 0 {
		t.Fatalf("provider invoked %d times during compilation", barrelSelfReturningProviderCalls)
	}
}
