package sail_test

import (
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/qlustra/intel/rules/typerule"
	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type recipeBarrelReader interface{ Next() int }
type recipeBarrelCounter struct {
	Count int
	Child bool `barreltest:"child" name:"nested"`
}

func (r *recipeBarrelCounter) Next() int { r.Count++; return r.Count }

type legendBarrelChild struct {
	legend.Landmark `sail:"landmark" family:"child"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}
type legendBarrelChildren struct {
	legend.Lane
	Child legendBarrelChild `sail:"spot" spot:"single,required"`
}
type legendBarrelRegion struct {
	legend.Region `sail:"region" family:"group"`
	Name          legend.Feature                         `sail:"feature" symbol:"name"`
	First         legend.Barrel[int, recipeBarrelReader] `sail:"barrel" populate:"Next"`
	Second        legend.Barrel[int, recipeBarrelReader] `sail:"barrel" populate:"Next"`
	Children      legendBarrelChildren                   `sail:"lane"`
}
type legendBarrelRegions struct {
	legend.Lane
	Groups []legendBarrelRegion `sail:"spot" spot:"multi,required"`
}
type legendBarrelRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	Groups        legendBarrelRegions `sail:"lane"`
}
type recipeBarrelGroups struct {
	A recipeBarrelCounter `barreltest:"group" name:"a"`
	B recipeBarrelCounter `barreltest:"group" name:"b"`
}

func TestPopulationBarrelRegionFreshSurrogates(t *testing.T) {
	chart, err := sail.Compile[legendBarrelRoot]()
	if err != nil {
		t.Fatal(err)
	}
	built, err := chart.Build[recipeBarrelGroups]()
	if err != nil {
		t.Fatal(err)
	}
	recipe := recipeBarrelGroups{A: recipeBarrelCounter{Count: 40}, B: recipeBarrelCounter{Count: 80}}
	populated, err := chart.Populate(recipe)
	if err != nil {
		t.Fatal(err)
	}
	again, err := chart.Build[recipeBarrelGroups]()
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []*legendBarrelRoot{built, populated, again} {
		if model == nil || len(model.Groups.Groups) != 2 {
			t.Fatalf("groups: %#v", model)
		}
		for i, group := range model.Groups.Groups {
			if !group.First.Present() || !group.Second.Present() || group.First.Get() != 1 || group.Second.Get() != 1 || group.Name.Get() != []string{"a", "b"}[i] || group.Children.Child.Name.Get() != "nested" {
				t.Fatalf("group %d: %#v", i, group)
			}
		}
	}
	if recipe.A.Count != 40 || recipe.B.Count != 80 {
		t.Fatal("live Recipe changed")
	}
	for range 4 {
		t.Run("concurrent fresh surrogates", func(t *testing.T) {
			t.Parallel()
			legend, err := chart.Build[recipeBarrelGroups]()
			if err != nil || legend == nil || len(legend.Groups.Groups) != 2 {
				t.Fatalf("concurrent Build: legend=%v error=%v", legend, err)
			}
			for _, group := range legend.Groups.Groups {
				if group.First.Get() != 1 || group.Second.Get() != 1 {
					t.Fatal("surrogate state leaked between captures or factory calls")
				}
			}
		})
	}
}

type recipeBarrelMissing struct{}
type recipeBarrelWrong struct{}

func (recipeBarrelWrong) Next() string { return "wrong" }

type legendOptionalDeclaration struct {
	legend.Landmark `sail:"landmark" family:"item"`
	Value           legend.Barrel[int, recipeBarrelReader] `sail:"barrel" populate:"Next"`
}
type legendRequiredDeclaration struct {
	legend.Landmark `sail:"landmark" family:"item"`
	Value           legend.Barrel[int, recipeBarrelReader] `sail:"barrel" populate:"Next" barrel:"required"`
}
type legendOptionalLane struct {
	legend.Lane
	Item legendOptionalDeclaration `sail:"spot" spot:"single"`
}
type legendRequiredLane struct {
	legend.Lane
	Item legendRequiredDeclaration `sail:"spot" spot:"single"`
}
type legendOptionalRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	Items         legendOptionalLane `sail:"lane"`
}
type legendRequiredRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	Items         legendRequiredLane `sail:"lane"`
}
type recipeBarrelGoodRoot struct {
	Item recipeBarrelCounter `barreltest:"item"`
}
type recipeBarrelMissingRoot struct {
	Item recipeBarrelMissing `barreltest:"item"`
}
type recipeBarrelWrongRoot struct {
	Item recipeBarrelWrong `barreltest:"item"`
}

func TestPopulationBarrelAvailabilityAndReuse(t *testing.T) {
	optional, err := sail.Compile[legendOptionalRoot]()
	if err != nil {
		t.Fatal(err)
	}
	good, err := optional.Build[recipeBarrelGoodRoot]()
	if err != nil || good == nil || !good.Items.Item.Value.Present() {
		t.Fatalf("good: %v", err)
	}
	missing, err := optional.Build[recipeBarrelMissingRoot]()
	if err != nil || missing == nil || missing.Items.Item.Value.Present() {
		t.Fatalf("missing: model=%v error=%v", missing, err)
	}
	wrong, err := optional.Build[recipeBarrelWrongRoot]()
	if err != nil || wrong == nil || wrong.Items.Item.Value.Present() {
		t.Fatalf("incompatible: model=%v error=%v", wrong, err)
	}
	again, err := optional.Build[recipeBarrelGoodRoot]()
	if err != nil || again == nil || again.Items.Item.Value.Get() != 1 || good.Items.Item.Value.Get() != 1 {
		t.Fatalf("reuse: %v", err)
	}
	required, err := sail.Compile[legendRequiredRoot]()
	if err != nil {
		t.Fatal(err)
	}
	if m, err := required.Build[recipeBarrelMissingRoot](); err == nil || m != nil || !strings.Contains(err.Error(), "Item") {
		t.Fatalf("required missing: model=%v error=%v", m, err)
	}
	if m, err := required.Build[recipeBarrelWrongRoot](); err == nil || m != nil || !strings.Contains(err.Error(), "Item") {
		t.Fatalf("required incompatible: model=%v error=%v", m, err)
	}
	if m, err := required.Build[recipeBarrelGoodRoot](); err != nil || m == nil || m.Items.Item.Value.Get() != 1 {
		t.Fatalf("after failures: %v", err)
	}
}

type recipeBarrelPrimitive int

type recipeBarrelDefaultable[T any] interface{ Default() T }

func (recipeBarrelPrimitive) Default() recipeBarrelPrimitive { return 7 }
func (recipeBarrelPrimitive) Other() string                  { return "different" }

type legendDynamicDeclaration struct {
	legend.Landmark `sail:"landmark" family:"item"`
	Default         legend.Barrel[any, any]                          `sail:"barrel" populate:"Default"`
	Other           legend.Barrel[any, any]                          `sail:"barrel" populate:"Other"`
	Literal         legend.Barrel[any, recipeBarrelDefaultable[any]] `sail:"barrel" populate:"Default"`
}
type legendDynamicLane struct {
	legend.Lane
	Item legendDynamicDeclaration `sail:"spot" spot:"single"`
}
type legendDynamicRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	Items         legendDynamicLane `sail:"lane"`
}
type RecipeBarrelPromoted struct {
	Original recipeBarrelPrimitive `barreltest:"item"`
}
type recipeBarrelPromoted struct{ RecipeBarrelPromoted }

func TestPopulationBarrelDynamicInspection(t *testing.T) {
	chart, err := sail.Compile[legendDynamicRoot]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[recipeBarrelPromoted]()
	if err != nil || m == nil {
		t.Fatalf("Build: %v", err)
	}
	value := m.Items.Item.Default
	if !value.Present() || value.Get() != recipeBarrelPrimitive(7) || m.Items.Item.Other.Get() != "different" {
		t.Fatal("dynamic values missing")
	}
	if m.Items.Item.Literal.Present() {
		t.Fatal("Default() named value must not satisfy literal Default() any")
	}
	method, check := value.Method()
	if !check.OK() || method == nil {
		t.Fatalf("method: %v", check)
	}
	resultType, check := method.Result(0)
	if !check.OK() || resultType.CapturedType() != reflect.TypeFor[recipeBarrelPrimitive]() {
		t.Fatalf("method result: %v", check)
	}
	result, check := value.Result()
	if !check.OK() || result.Type().CapturedType() != reflect.TypeFor[recipeBarrelPrimitive]() || result.Get() != recipeBarrelPrimitive(7) {
		t.Fatalf("result: %v", check)
	}
	carrier, check := value.CarrierType()
	if !check.OK() || carrier.CapturedType() != reflect.TypeFor[recipeBarrelPrimitive]() {
		t.Fatalf("carrier: %v", check)
	}
}

type recipeZeroCarrier int

func (recipeZeroCarrier) Empty() string { return "" }
func (recipeZeroCarrier) Zero() int     { return 0 }
func (recipeZeroCarrier) Nil() any      { return nil }
func (recipeZeroCarrier) Next() int     { return 9 }

type recipePanicCarrier int

func (recipePanicCarrier) Empty() string { return "" }
func (recipePanicCarrier) Zero() int     { return 0 }
func (recipePanicCarrier) Nil() any      { return nil }
func (recipePanicCarrier) Next() int     { panic("getter failed") }

type LegendBarrelValues struct {
	Empty legend.Barrel[string, any] `sail:"barrel" populate:"Empty"`
	Zero  legend.Barrel[int, any]    `sail:"barrel" populate:"Zero"`
	Nil   legend.Barrel[any, any]    `sail:"barrel" populate:"Nil"`
	Next  legend.Barrel[int, any]    `sail:"barrel" populate:"Next"`
}
type legendValuesDeclaration struct {
	legend.Landmark `sail:"landmark" family:"item"`
	LegendBarrelValues
}
type legendValuesLane struct {
	legend.Lane
	Items []legendValuesDeclaration `sail:"spot" spot:"multi"`
}
type legendValuesRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	Items         legendValuesLane        `sail:"lane" exploration:"eager"`
	Fallback      legendGuardFallbackLane `sail:"lane"`
}
type recipeValuesGood struct {
	First recipeZeroCarrier `barreltest:"item"`
}
type recipeValuesBad struct {
	First recipeZeroCarrier  `barreltest:"item"`
	Later recipePanicCarrier `barreltest:"item"`
}

func TestPopulationBarrelZeroAndFatalPanic(t *testing.T) {
	chart, err := sail.Compile[legendValuesRoot]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[recipeValuesGood]()
	if err != nil || m == nil || len(m.Items.Items) != 1 {
		t.Fatalf("good: %v", err)
	}
	v := m.Items.Items[0].LegendBarrelValues
	if !v.Empty.Present() || v.Empty.Get() != "" || !v.Zero.Present() || v.Zero.Get() != 0 || !v.Nil.Present() || v.Nil.Get() != nil || !v.Next.Present() || v.Next.Get() != 9 {
		t.Fatal("zero results not present")
	}
	if failed, err := chart.Build[recipeValuesBad](); err == nil || failed != nil || !strings.Contains(err.Error(), "Later") {
		t.Fatalf("panic: model=%v error=%v", failed, err)
	}
	if m.Items.Items[0].Next.Get() != 9 || !m.Items.Items[0].Nil.Present() {
		t.Fatal("failure changed earlier Legend")
	}
}

type recipeNilPointerCarrier string

func (recipeNilPointerCarrier) Nil() *int { return nil }
func (recipeNilPointerCarrier) Next() int { return 23 }

type recipeBarrelMetadata struct {
	A recipeZeroCarrier       `barreltest:"item"`
	B recipeNilPointerCarrier `barreltest:"item"`
	C recipeZeroCarrier       `barreltest:"item"`
	D recipeNilPointerCarrier `barreltest:"item"`
}

func TestPopulationBarrelInspectionSurvivesSliceGrowth(t *testing.T) {
	chart, err := sail.Compile[legendValuesRoot]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[recipeBarrelMetadata]()
	if err != nil || populatedLegend == nil || len(populatedLegend.Items.Items) != 4 {
		t.Fatalf("Build: legend=%v error=%v", populatedLegend, err)
	}
	for index, occurrence := range populatedLegend.Items.Items {
		expectedCarrier := reflect.TypeFor[recipeZeroCarrier]()
		expectedResult := reflect.TypeFor[any]()
		if index%2 == 1 {
			expectedCarrier = reflect.TypeFor[recipeNilPointerCarrier]()
			expectedResult = reflect.TypeFor[*int]()
			if pointer, ok := occurrence.Nil.Get().(*int); !ok || pointer != nil {
				t.Fatal("typed nil result was not retained in the assigned view")
			}
			if occurrence.Empty.Present() || occurrence.Zero.Present() || occurrence.Next.Get() != 23 {
				t.Fatal("optional state leaked from a previous occurrence")
			}
		} else if occurrence.Nil.Get() != nil || occurrence.Next.Get() != 9 {
			t.Fatal("nil interface result or value was not retained")
		}
		if !occurrence.Nil.Present() {
			t.Fatal("successful nil capture is absent")
		}
		result, check := occurrence.Nil.Result()
		if !check.OK() || result.Type().CapturedType() != expectedResult {
			t.Fatalf("occurrence %d result inspection: %v", index, check)
		}
		method, check := occurrence.Nil.Method()
		if !check.OK() {
			t.Fatal(check.Why())
		}
		methodResult, check := method.Result(0)
		if !check.OK() || methodResult.CapturedType() != expectedResult {
			t.Fatalf("occurrence %d method result inspection: %v", index, check)
		}
		carrier, check := occurrence.Nil.CarrierType()
		if !check.OK() || carrier.CapturedType() != expectedCarrier {
			t.Fatalf("occurrence %d carrier inspection: %v", index, check)
		}
	}
}

var recipeBarrelInvocations atomic.Int64

type recipeGuardCarrier int

func (recipeGuardCarrier) Next() int { recipeBarrelInvocations.Add(1); return 1 }

type legendGuardDeclaration struct {
	legend.Landmark `sail:"landmark" family:"item"`
	Value           legend.Barrel[int, any] `sail:"barrel" populate:"Next"`
}

func (legendGuardDeclaration) Accept() typerule.RuleAPI { return typerule.Type().Is[bool]() }

type legendGuardLane struct {
	legend.Lane
	Item legendGuardDeclaration `sail:"spot" spot:"single"`
}
type legendGuardFallback struct {
	legend.Landmark `sail:"landmark" family:"item"`
}
type legendGuardFallbackLane struct {
	legend.Lane
	Item legendGuardFallback `sail:"spot" spot:"single"`
}
type legendGuardRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	First         legendGuardLane         `sail:"lane" exploration:"eager"`
	Fallback      legendGuardFallbackLane `sail:"lane"`
}
type recipeGuardRoot struct {
	Rejected recipeGuardCarrier `barreltest:"item"`
}

type recipeInvalidRegionRoot struct {
	Group recipeGuardCarrier `barreltest:"group"`
}

func TestPopulationBarrelRegionAdmissionBeforeProvider(t *testing.T) {
	recipeBarrelInvocations.Store(0)
	chart, err := sail.Compile[legendBarrelRoot]()
	if err != nil {
		t.Fatal(err)
	}
	// The named primitive has the requested getter, but cannot supply a Region
	// scope. Reject its shape before declaration-level provider execution.
	if legend, err := chart.Build[recipeInvalidRegionRoot](); err == nil || legend != nil || !strings.Contains(err.Error(), "struct value") {
		t.Fatalf("invalid Region carrier: legend=%v error=%v", legend, err)
	}
	if recipeBarrelInvocations.Load() != 0 {
		t.Fatal("inadmissible Region carrier invoked a provider")
	}
}

func TestPopulationBarrelAcceptanceBeforeProvider(t *testing.T) {
	recipeBarrelInvocations.Store(0)
	chart, err := sail.Compile[legendGuardRoot]()
	if err != nil {
		t.Fatal(err)
	}
	if recipeBarrelInvocations.Load() != 0 {
		t.Fatal("Compile invoked Recipe getter")
	}
	if m, err := chart.Build[recipeGuardRoot](); err == nil || m != nil || !strings.Contains(err.Error(), "acceptance") {
		t.Fatalf("rejected: model=%v error=%v", m, err)
	}
	if recipeBarrelInvocations.Load() != 0 {
		t.Fatal("rejected getter invoked")
	}
}

type legendBarrelQualifier struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[int, any] `sail:"barrel" populate:"Next" barrel:"required"`
}
type legendQualifierDeclaration struct {
	legend.Landmark `sail:"landmark" family:"item"`
	Enabled         legendBarrelQualifier `sail:"qualifier" symbol:"enabled"`
}
type legendQualifierLane struct {
	legend.Lane
	Item legendQualifierDeclaration `sail:"spot" spot:"single"`
}
type legendQualifierRoot struct {
	legend.Legend `sail:"legend" mark:"barreltest"`
	Items         legendQualifierLane `sail:"lane"`
}
type recipeQualifierBare struct {
	Item recipeGuardCarrier `barreltest:"item"`
}
type recipeQualifierToken struct {
	Item recipeGuardCarrier `barreltest:"item,enabled"`
}

func TestPopulationBarrelQualifierGating(t *testing.T) {
	recipeBarrelInvocations.Store(0)
	chart, err := sail.Compile[legendQualifierRoot]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[recipeQualifierBare]()
	if err != nil || m == nil {
		t.Fatalf("bare: %v", err)
	}
	if m.Items.Item.Enabled.Present() || m.Items.Item.Enabled.Value.Present() || recipeBarrelInvocations.Load() != 0 {
		t.Fatal("inactive qualifier invoked provider or became present")
	}
	activeLegend, err := chart.Build[recipeQualifierToken]()
	if err != nil || activeLegend == nil || !activeLegend.Items.Item.Enabled.Present() || !activeLegend.Items.Item.Enabled.Value.Present() || activeLegend.Items.Item.Enabled.Value.Get() != 1 {
		t.Fatalf("active qualifier: legend=%v error=%v", activeLegend, err)
	}
	if recipeBarrelInvocations.Load() != 1 {
		t.Fatal("active qualifier did not invoke its provider exactly once")
	}
}
