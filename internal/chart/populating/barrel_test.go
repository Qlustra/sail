package populating

import (
	"reflect"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/carriers"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/definitions"
	"github.com/qlustra/sail/legend"
)

type barrelReader interface{ Read() int }
type Defaultable[T any] interface{ Default() T }
type namedInt int
type namedRecord struct{ N int }
type remote struct{ Name string }

func (remote) String() string { return "remote" }

type barrelFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[int, barrelReader] `sail:"barrel" populate:"Read"`
}
type remoteDynamicFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[any, any] `sail:"barrel" populate:"Default"`
}
type remoteLiteralFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[any, Defaultable[any]] `sail:"barrel" populate:"Default" barrel:"required"`
}
type interfaceResultFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[barrelReader, any] `sail:"barrel" populate:"Value"`
}
type multiBarrelFixture struct {
	legend.Region `sail:"region" family:"test"`
	First         legend.Barrel[int, any] `sail:"barrel" populate:"Next"`
	Second        legend.Barrel[int, any] `sail:"barrel" populate:"Next"`
}
type unnamedRecordFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[namedRecord, any] `sail:"barrel" populate:"Read"`
}
type namedRecordFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[struct{ N int }, any] `sail:"barrel" populate:"Read"`
}
type requiredBarrelFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[int, barrelReader] `sail:"barrel" populate:"Read" barrel:"required"`
}
type dynamicBarrelFixture struct {
	legend.Region `sail:"region" family:"test"`
	Value         legend.Barrel[int, any] `sail:"barrel" populate:"Read"`
}
type barrelProvider struct{}

func (*barrelProvider) Read() int { return 42 }

type wrongBarrelProvider struct{}

var wrongMethodCalls int

func (*wrongBarrelProvider) Read() string { wrongMethodCalls++; return "wrong" }

type panicBarrelProvider struct{}

func (*panicBarrelProvider) Read() int { panic("provider") }

type noBarrelProvider struct{}

func (remote) Default() remote      { return remote{Name: "dynamic"} }
func (remote) WrongDefault() string { return "wrong" }

type anyDefaultProvider struct{}

func (anyDefaultProvider) Default() any { return remote{Name: "literal"} }

type interfaceNilProvider struct{}

func (*interfaceNilProvider) Value() barrelReader { return nil }

type interfaceTypedNilProvider struct{}

func (*interfaceTypedNilProvider) Value() barrelReader { var p *barrelProvider; return p }

type counterProvider struct{ count int }

var counterProviderCalls int

func (p *counterProvider) Next() int { counterProviderCalls++; p.count++; return p.count }

type nilBarrelHolder struct{ *legend.Barrel[int, any] }

type recordProvider struct{}

func (recordProvider) Read() struct{ N int } { return struct{ N int }{N: 5} }

type namedRecordProvider struct{}

func (namedRecordProvider) Read() namedRecord { return namedRecord{N: 6} }

type namedCarrier int

type P *P

func (*namedCarrier) Read() int { return int(*new(namedCarrier)) + 7 }

func TestCaptureBarrelPointerCycleIsUnavailable(t *testing.T) {
	binding := compiledBarrel(t, dynamicBarrelFixture{}).Barrels[0]
	var destination legend.Barrel[int, any]
	if check := CaptureBarrel(binding, capture.Type[P](), &destination); !check.OK() {
		t.Fatalf("optional pointer cycle should be unavailable: %v", check)
	}
}

func compiledBarrel(t *testing.T, source any) *declarations.Region {
	t.Helper()
	d, err := define.Struct[definitions.RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.TypeOf(source)
	if !check.OK() {
		t.Fatal(check)
	}
	structType, check := subject.AsStruct()
	if !check.OK() {
		t.Fatal(check)
	}
	ctx, validation := d.Compile(structType, nil)
	if !validation.IsOkay() {
		t.Fatal(validation.Trace())
	}
	region, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check)
	}
	return region
}

func TestCaptureBarrelFromCompiledBinding(t *testing.T) {
	region := compiledBarrel(t, barrelFixture{})
	binding := region.Barrels[0]
	var destination legend.Barrel[int, barrelReader]
	check := CaptureBarrel(binding, capture.Type[barrelProvider](), &destination)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if !destination.Present() || destination.Get() != 42 {
		t.Fatalf("captured value = %d, present=%v", destination.Get(), destination.Present())
	}
	if _, check := destination.Method(); !check.OK() {
		t.Fatal(check)
	}
	if result, check := destination.Result(); !check.OK() || result.Get() != 42 {
		t.Fatalf("result: %v", check)
	}
	if carrier, check := destination.CarrierType(); !check.OK() || carrier.CapturedType() != reflect.TypeFor[barrelProvider]() {
		t.Fatalf("carrier type: %v", check)
	}
}

func TestCaptureBarrelChecksEmbeddedDestinationBeforeInvocation(t *testing.T) {
	binding := compiledBarrel(t, dynamicBarrelFixture{}).Barrels[0]
	counterProviderCalls = 0
	if check := CaptureBarrel(binding, capture.Type[counterProvider](), &nilBarrelHolder{}); check.OK() {
		t.Fatal("accepted nil embedded Barrel")
	}
	if counterProviderCalls != 0 {
		t.Fatalf("provider invoked %d times before destination readiness", counterProviderCalls)
	}
}

func TestCaptureBarrelAvailabilityAndDestinationValidation(t *testing.T) {
	optional := compiledBarrel(t, barrelFixture{}).Barrels[0]
	var destination legend.Barrel[int, barrelReader]
	for _, carrier := range []*capture.TypeInterrogator{capture.Type[noBarrelProvider](), capture.Type[wrongBarrelProvider]()} {
		if check := CaptureBarrel(optional, carrier, &destination); !check.OK() {
			t.Fatalf("optional absence: %s", check.Why())
		}
		if destination.Present() {
			t.Fatal("optional absence wrote a value")
		}
	}
	required := compiledBarrel(t, requiredBarrelFixture{}).Barrels[0]
	if check := CaptureBarrel(required, capture.Type[noBarrelProvider](), &destination); check.OK() || check.Code != "barrel_unavailable" {
		t.Fatalf("required absence: %v", check)
	}
	if check := CaptureBarrel(optional, capture.Type[barrelProvider](), new(int)); check.OK() || check.Code != "invalid_barrel_destination" {
		t.Fatalf("invalid destination: %v", check)
	}
	if check := CaptureBarrel(optional, capture.Type[noBarrelProvider](), new(int)); check.OK() || check.Code != "invalid_barrel_destination" {
		t.Fatalf("destination must precede availability: %v", check)
	}
	if check := CaptureBarrel(optional, capture.Type[barrelProvider](), nil); check.OK() {
		t.Fatal("accepted nil destination")
	}
	if check := CaptureBarrel(optional, capture.Type[any](), &destination); !check.OK() {
		t.Fatalf("optional interface receiver absence: %v", check)
	}
	if check := CaptureBarrel(compiledBarrel(t, requiredBarrelFixture{}).Barrels[0], capture.Type[any](), &destination); check.OK() || check.Code != "barrel_unavailable" {
		t.Fatalf("required interface receiver absence: %v", check)
	}
	optional.Shape.Capability = nil
	if check := CaptureBarrel(optional, nil, &destination); check.OK() || check.Code != "invalid_barrel_capture" {
		t.Fatalf("nil shape metadata must fail independently of optionality: %v", check)
	}
	optional.Shape.Capability = &capture.TypeInterrogator{}
	if check := CaptureBarrel(optional, nil, &destination); check.OK() || check.Code != "invalid_barrel_capture" {
		t.Fatalf("nil captured type must fail independently of optionality: %v", check)
	}
}

func TestCaptureBarrelInvocationFailureIsFatal(t *testing.T) {
	optional := compiledBarrel(t, barrelFixture{}).Barrels[0]
	var destination legend.Barrel[int, barrelReader]
	if check := CaptureBarrel(optional, capture.Type[panicBarrelProvider](), &destination); check.OK() || check.Code != "barrel_invocation_failed" {
		t.Fatalf("provider panic: %v", check)
	}
	if destination.Present() {
		t.Fatal("failed invocation mutated destination")
	}
}

func TestCaptureBarrelDynamicSignatureMismatchIsUnavailable(t *testing.T) {
	optional := compiledBarrel(t, dynamicBarrelFixture{}).Barrels[0]
	var destination legend.Barrel[int, any]
	wrongMethodCalls = 0
	if check := CaptureBarrel(optional, capture.Type[wrongBarrelProvider](), &destination); !check.OK() {
		t.Fatalf("optional signature mismatch should be absent: %s", check.Why())
	}
	if destination.Present() {
		t.Fatal("signature mismatch wrote a value")
	}
	if wrongMethodCalls != 0 {
		t.Fatalf("mismatched method invoked %d times", wrongMethodCalls)
	}
	optional.Required = true
	if check := CaptureBarrel(optional, capture.Type[wrongBarrelProvider](), &destination); check.OK() || check.Code != "barrel_unavailable" {
		t.Fatalf("required signature mismatch: %v", check)
	}
}

func TestCaptureBarrelDynamicRemoteAndLiteralCapability(t *testing.T) {
	dynamic := compiledBarrel(t, remoteDynamicFixture{}).Barrels[0]
	var dynamicValue legend.Barrel[any, any]
	if check := CaptureBarrel(dynamic, capture.Type[remote](), &dynamicValue); !check.OK() {
		t.Fatal(check)
	}
	if got, ok := dynamicValue.Get().(remote); !ok || got.Name != "dynamic" {
		t.Fatalf("dynamic result = %#v", dynamicValue.Get())
	}
	dynamicResult, check := dynamicValue.Result()
	if !check.OK() {
		t.Fatalf("dynamic result inspection: %v", check)
	}
	if declared := dynamic.Shape.Result.CapturedType(); declared != reflect.TypeFor[any]() {
		t.Fatalf("declared result type = %v, want any", declared)
	}
	if runtime := dynamicResult.Type().CapturedType(); runtime != reflect.TypeFor[remote]() {
		t.Fatalf("runtime result type = %v, want remote", runtime)
	}

	literal := compiledBarrel(t, remoteLiteralFixture{}).Barrels[0]
	if check := CaptureBarrel(literal, capture.Type[remote](), new(legend.Barrel[any, Defaultable[any]])); check.OK() || check.Code != "barrel_unavailable" {
		t.Fatalf("Default() Remote must not satisfy Default() any: %v", check)
	}
	var literalValue legend.Barrel[any, Defaultable[any]]
	if check := CaptureBarrel(literal, capture.Type[anyDefaultProvider](), &literalValue); !check.OK() {
		t.Fatal(check)
	}
	if got, ok := literalValue.Get().(remote); !ok || got.Name != "literal" {
		t.Fatalf("literal result = %#v", literalValue.Get())
	}
}

func TestCaptureBarrelDeclaredInterfaceAndRuntimeNilValues(t *testing.T) {
	binding := compiledBarrel(t, interfaceResultFixture{}).Barrels[0]
	for _, tc := range []struct {
		name         string
		carrier      *capture.TypeInterrogator
		wantNil      bool
		wantTypedNil bool
	}{
		{"nil interface", capture.Type[interfaceNilProvider](), true, false},
		{"typed nil interface value", capture.Type[interfaceTypedNilProvider](), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var destination legend.Barrel[barrelReader, any]
			if check := CaptureBarrel(binding, tc.carrier, &destination); !check.OK() {
				t.Fatal(check)
			}
			result, check := destination.Result()
			if !check.OK() || result.Type().CapturedType() != reflect.TypeFor[barrelReader]() {
				t.Fatalf("declared result type: %v", check)
			}
			value := destination.Get()
			if tc.wantNil && value != nil {
				t.Fatalf("got %#v, want nil interface", value)
			}
			if tc.wantNil && result.Get() != nil {
				t.Fatalf("result is %#v, want nil interface", result.Get())
			}
			if tc.wantTypedNil {
				p, ok := value.(*barrelProvider)
				if !ok || p != nil {
					t.Fatalf("got %#v, want typed nil *barrelProvider", value)
				}
				provider, ok := result.Get().(*barrelProvider)
				if !ok || provider != nil {
					t.Fatalf("raw result = %#v, want typed nil *barrelProvider", result.Get())
				}
			}
		})
	}
}

func TestCaptureBarrelNamedPrimitivePointerCarrier(t *testing.T) {
	binding := compiledBarrel(t, dynamicBarrelFixture{}).Barrels[0]
	var destination legend.Barrel[int, any]
	if check := CaptureBarrel(binding, capture.Type[namedCarrier](), &destination); !check.OK() {
		t.Fatal(check)
	}
	if destination.Get() != 7 {
		t.Fatalf("captured %d", destination.Get())
	}
}

func TestCaptureBarrelUsesFreshReceiverForDistinctOccurrences(t *testing.T) {
	region := compiledBarrel(t, multiBarrelFixture{})
	if len(region.Barrels) != 2 || region.Barrels[0] == region.Barrels[1] {
		t.Fatalf("barrels: %#v", region.Barrels)
	}
	var first, second legend.Barrel[int, any]
	counterProviderCalls = 0
	for i, tc := range []struct {
		binding     *carriers.Barrel
		destination any
	}{
		{region.Barrels[0], &first}, {region.Barrels[1], &second},
	} {
		if check := CaptureBarrel(tc.binding, capture.Type[counterProvider](), tc.destination); !check.OK() {
			t.Fatalf("capture %d: %v", i, check)
		}
	}
	if first.Get() != 1 || second.Get() != 1 {
		t.Fatalf("results %d, %d; expected fresh zero receiver per occurrence", first.Get(), second.Get())
	}
	if counterProviderCalls != 2 {
		t.Fatalf("provider invoked %d times for two captures", counterProviderCalls)
	}
	firstResult, check := first.Result()
	if !check.OK() {
		t.Fatalf("first runtime result accessor: %v", check)
	}
	secondResult, check := second.Result()
	if !check.OK() {
		t.Fatalf("second runtime result accessor: %v", check)
	}
	if firstResult == secondResult {
		t.Fatal("occurrences share result interrogation")
	}
}

func TestCaptureBarrelNamedAndUnnamedAssignment(t *testing.T) {
	for _, tc := range []struct {
		source     any
		carrier    reflect.Type
		want       int
		valueType  reflect.Type
		resultType reflect.Type
	}{
		{unnamedRecordFixture{}, reflect.TypeFor[recordProvider](), 5, reflect.TypeFor[namedRecord](), reflect.TypeFor[struct{ N int }]()},
		{namedRecordFixture{}, reflect.TypeFor[namedRecordProvider](), 6, reflect.TypeFor[struct{ N int }](), reflect.TypeFor[namedRecord]()},
	} {
		binding := compiledBarrel(t, tc.source).Barrels[0]
		if reflect.TypeOf(tc.source) == reflect.TypeFor[unnamedRecordFixture]() {
			var value legend.Barrel[namedRecord, any]
			if check := CaptureBarrel(binding, capture.TypeReflection(tc.carrier), &value); !check.OK() {
				t.Fatal(check)
			}
			if value.Get().N != tc.want {
				t.Fatalf("value=%v", value.Get())
			}
			result, check := value.Result()
			if !check.OK() {
				t.Fatalf("unnamed assignment result accessor: %v", check)
			}
			if reflect.TypeOf(value.Get()) != tc.valueType || result.Type().CapturedType() != tc.resultType {
				t.Fatalf("assignment view=%T, raw result type=%v want %v", value.Get(), result.Type().CapturedType(), tc.resultType)
			}
		} else {
			var value legend.Barrel[struct{ N int }, any]
			if check := CaptureBarrel(binding, capture.TypeReflection(tc.carrier), &value); !check.OK() {
				t.Fatal(check)
			}
			if value.Get().N != tc.want {
				t.Fatalf("value=%v", value.Get())
			}
			result, check := value.Result()
			if !check.OK() {
				t.Fatalf("named assignment result accessor: %v", check)
			}
			if reflect.TypeOf(value.Get()) != tc.valueType || result.Type().CapturedType() != tc.resultType {
				t.Fatalf("assignment view=%T, raw result type=%v want %v", value.Get(), result.Type().CapturedType(), tc.resultType)
			}
		}
	}
}

func TestBarrelInspectionStartsAbsent(t *testing.T) {
	var value legend.Barrel[int, any]
	if _, check := value.Method(); check.OK() {
		t.Fatal("Method available before capture")
	}
	if _, check := value.Result(); check.OK() {
		t.Fatal("Result available before capture")
	}
	if _, check := value.CarrierType(); check.OK() {
		t.Fatal("CarrierType available before capture")
	}
}

func TestCaptureDoesNotMutateCompiledMetadata(t *testing.T) {
	region := compiledBarrel(t, barrelFixture{})
	binding := region.Barrels[0]
	metadata := *binding
	trace := region.Trace()
	var destination legend.Barrel[int, barrelReader]
	if check := CaptureBarrel(binding, capture.Type[barrelProvider](), &destination); !check.OK() {
		t.Fatal(check)
	}
	if !reflect.DeepEqual(*binding, metadata) {
		t.Fatal("capture mutated compiled Barrel metadata")
	}
	if after := region.Trace(); after != trace {
		t.Fatalf("capture changed declaration trace:\nbefore:\n%s\nafter:\n%s", trace, after)
	}
}

func TestCaptureBarrelWriteFailureIsFatal(t *testing.T) {
	optional := compiledBarrel(t, barrelFixture{}).Barrels[0]
	var destination legend.Barrel[string, barrelReader]
	if check := CaptureBarrel(optional, capture.Type[barrelProvider](), &destination); check.OK() || check.Code != "barrel_write_failed" {
		t.Fatalf("write failure: %v", check)
	}
	if destination.Present() {
		t.Fatal("failed write mutated destination")
	}
}
