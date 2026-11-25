package main

import (
	"reflect"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type declaredLabel string

type labelProvider struct{}

func (labelProvider) Label() declaredLabel { return "compile" }

type describedValue struct{ legend.Barrel[any, any] }

type describedJob struct {
	legend.Landmark `sail:"landmark" family:"job"`
	Label           describedValue `sail:"barrel" populate:"Label"`
}

type describedJobs struct {
	legend.Lane
	Jobs []describedJob `sail:"spot" spot:"multi"`
}

type describedPipeline struct {
	legend.Legend `sail:"legend" mark:"pipeline"`
	Jobs          describedJobs `sail:"lane"`
}

func TestBarrelReadersWithoutPublicShapeHook(t *testing.T) {
	if _, exposed := reflect.TypeFor[*legend.Barrel[any, any]]().MethodByName("BarrelShape"); exposed {
		t.Fatal("Barrel exposes a compiler-only shape hook")
	}
	navigator, err := sail.Compile[describedPipeline]()
	if err != nil {
		t.Fatal(err)
	}
	type chart struct {
		Compile labelProvider `pipeline:"job"`
		Other   struct{}      `pipeline:"job"`
	}
	model, err := navigator.Build[chart]()
	if err != nil {
		t.Fatal(err)
	}
	captured := model.Jobs.Jobs[0].Label
	if !captured.Present() || captured.Get() != declaredLabel("compile") {
		t.Fatalf("present=%t value=%v", captured.Present(), captured.Get())
	}
	if method, check := captured.Method(); !check.OK() || method == nil {
		t.Fatalf("Method() failed: %v", check)
	}
	if result, check := captured.Result(); !check.OK() || result.Type().CapturedType() != reflect.TypeFor[declaredLabel]() || result.Get() != declaredLabel("compile") {
		t.Fatalf("unexpected captured result: %v (%v)", result, check)
	}
	if carrier, check := captured.CarrierType(); !check.OK() || carrier.CapturedType() != reflect.TypeFor[labelProvider]() {
		t.Fatalf("unexpected carrier type: %v (%v)", carrier, check)
	}
	absent := model.Jobs.Jobs[1].Label
	if _, check := absent.Result(); absent.Present() || check.OK() {
		t.Fatalf("optional Barrel without a provider: present=%t, result check=%v", absent.Present(), check)
	}
}
