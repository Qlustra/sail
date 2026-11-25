package populating

import (
	"reflect"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/legend"
)

func TestCaptureBarrelKeepsDeclarationSeparateFromRuntimeInspection(t *testing.T) {
	region := compiledBarrel(t, barrelFixture{})
	if len(region.Barrels) != 1 {
		t.Fatalf("compiled barrels = %d, want 1", len(region.Barrels))
	}
	binding := region.Barrels[0]
	metadata := *binding
	var destination legend.Barrel[int, barrelReader]
	if check := CaptureBarrel(binding, capture.Type[barrelProvider](), &destination); !check.OK() {
		t.Fatal(check)
	}

	if !reflect.DeepEqual(*binding, metadata) {
		t.Fatal("runtime capture changed compiled declaration metadata")
	}
	if destination.Get() != 42 || !destination.Present() {
		t.Fatalf("runtime result = %d, present=%v", destination.Get(), destination.Present())
	}
	runtimeMethod, check := destination.Method()
	if !check.OK() {
		t.Fatalf("runtime method accessor: %v", check)
	}
	if runtimeMethod == nil || binding.Method == nil {
		t.Fatal("expected both declaration method and runtime method inspection")
	}
	runtimeResult, check := destination.Result()
	if !check.OK() {
		t.Fatalf("runtime result accessor: %v", check)
	}
	if runtimeResult == nil || runtimeResult.Get() != 42 {
		t.Fatalf("runtime result inspection = %#v", runtimeResult)
	}
	runtimeCarrier, check := destination.CarrierType()
	if !check.OK() {
		t.Fatalf("runtime carrier accessor: %v", check)
	}
	if runtimeCarrier == nil || runtimeCarrier.CapturedType() != reflect.TypeFor[barrelProvider]() {
		t.Fatalf("runtime carrier type = %v", runtimeCarrier)
	}
	if binding.Shape.Result.CapturedType() != reflect.TypeFor[int]() || binding.Shape.Capability.CapturedType() != reflect.TypeFor[barrelReader]() {
		t.Fatal("compiled declaration result/capability types changed")
	}
}

func TestBarrelRuntimeAccessorsReportAbsentInspection(t *testing.T) {
	var value legend.Barrel[int, any]
	if _, check := value.Method(); check.OK() {
		t.Fatal("Method accessor succeeded before runtime capture")
	}
	if _, check := value.Result(); check.OK() {
		t.Fatal("Result accessor succeeded before runtime capture")
	}
	if _, check := value.CarrierType(); check.OK() {
		t.Fatal("CarrierType accessor succeeded before runtime capture")
	}
}

type differentResultProvider struct{}

func (differentResultProvider) Default() string { return "not the carrier type" }

func TestDynamicBarrelDoesNotEnforceSelfPolicy(t *testing.T) {
	binding := compiledBarrel(t, remoteDynamicFixture{}).Barrels[0]
	var destination legend.Barrel[any, any]
	if check := CaptureBarrel(binding, capture.Type[differentResultProvider](), &destination); !check.OK() {
		t.Fatal(check.Why())
	}
	if destination.Get() != "not the carrier type" || !destination.Present() {
		t.Fatal("dynamic Barrel did not retain its valid non-self result")
	}
	method, check := destination.Method()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	resultType, check := method.Result(0)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	carrierType, check := destination.CarrierType()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if resultType.Equals(carrierType.CapturedType()).OK() {
		t.Fatal("usage-time assertion unexpectedly equated string and carrier type")
	}
}
