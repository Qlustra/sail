package values_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/carriers"
	"github.com/qlustra/sail/internal/chart/values"
	"github.com/qlustra/sail/legend"
)

type reader[T any] interface {
	Get() T
	Present() bool
}

type inspectionProvider struct{}

func (*inspectionProvider) Read() int { return 0 }

func TestBarrelInspectionAndCombinedCapture(t *testing.T) {
	method, check := capture.Type[inspectionProvider]().ZeroReceiver()
	if !check.OK() {
		t.Fatal(check)
	}
	getterMethod, check := method.Method("Read")
	if !check.OK() {
		t.Fatal(check)
	}
	getter, check := method.Getter(getterMethod)
	if !check.OK() {
		t.Fatal(check)
	}
	result, check := getter.Read()
	if !check.OK() {
		t.Fatal(check)
	}
	var destination legend.Barrel[int, any]
	if check := values.CaptureBarrel(&destination, 0, getterMethod, result, capture.Type[inspectionProvider]()); !check.OK() {
		t.Fatal(check)
	}
	if !destination.Present() || destination.Get() != 0 {
		t.Fatal("zero result not captured")
	}
	if got, check := destination.Method(); !check.OK() || got == nil {
		t.Fatalf("Method: %v", check)
	}
	if got, check := destination.Result(); !check.OK() || got.Get() != 0 {
		t.Fatalf("Result: %v", check)
	}
	if got, check := destination.CarrierType(); !check.OK() || got.CapturedType() != reflect.TypeFor[inspectionProvider]() {
		t.Fatalf("CarrierType: %v", check)
	}
}

var _ reader[string] = legend.Feature{}
var _ reader[int] = legend.Barrel[int, any]{}
var _ interface{ Present() bool } = legend.Qualifier{}

type named string
type namedSlice []int
type namedMap map[string]int
type namedStruct struct{ N int }
type custom struct{ legend.Qualifier }
type broken struct{ *legend.Qualifier }

func TestOccurrenceValues(t *testing.T) {
	var a, b legend.Feature
	if a.Present() || a.Get() != "" {
		t.Fatal("feature zero")
	}
	if !values.Fill(&a, "hello").OK() || a.Get() != "hello" || !a.Present() || b.Present() {
		t.Fatal("feature fill or independence")
	}
	if values.Fill(&a, 1).OK() || a.Get() != "hello" || !a.Present() {
		t.Fatal("wrong type changed state")
	}
	if !values.Fill(&a, "").OK() || !a.Present() || a.Get() != "" {
		t.Fatal("empty capture")
	}
	var q legend.Barrel[int, any]
	if q.Present() || q.Get() != 0 || !values.Fill(&q, 0).OK() || !q.Present() {
		t.Fatal("zero capture")
	}
	var n legend.Barrel[named, any]
	if values.Fill(&n, "name").OK() || n.Present() {
		t.Fatal("accepted distinct type")
	}
	if !values.Fill(&n, named("name")).OK() || n.Get() != named("name") {
		t.Fatal("named capture")
	}
	var p legend.Barrel[*int, any]
	if !values.Fill(&p, nil).OK() || !p.Present() || p.Get() != nil {
		t.Fatal("nil pointer capture")
	}
	var typed *int
	var x legend.Barrel[any, any]
	if !values.Fill(&x, typed).OK() || !x.Present() || x.Get() != typed {
		t.Fatal("typed nil into any")
	}
	if !values.Fill(&x, nil).OK() || x.Get() != nil || !x.Present() {
		t.Fatal("nil into any")
	}
	var u legend.Barrel[unsafe.Pointer, any]
	if !values.Fill(&u, nil).OK() || !u.Present() || u.Get() != nil {
		t.Fatal("nil unsafe pointer")
	}
	if values.Fill(&q, nil).OK() || !q.Present() || q.Get() != 0 {
		t.Fatal("invalid nil changed state")
	}
}

func TestAssignableCaptures(t *testing.T) {
	var slice legend.Barrel[namedSlice, any]
	source := []int{7}
	if c := values.Fill(&slice, source); !c.OK() || !slice.Present() || slice.Get()[0] != 7 {
		t.Fatal("slice", c)
	}
	if values.Fill(&slice, []string{"wrong"}).OK() || !slice.Present() || slice.Get()[0] != 7 {
		t.Fatal("bad write mutated slice")
	}
	var mapping legend.Barrel[namedMap, any]
	if c := values.Fill(&mapping, map[string]int{"x": 8}); !c.OK() || !mapping.Present() || mapping.Get()["x"] != 8 {
		t.Fatal("map", c)
	}
	var structure legend.Barrel[namedStruct, any]
	if c := values.Fill(&structure, struct{ N int }{9}); !c.OK() || structure.Get().N != 9 || !structure.Present() {
		t.Fatal("struct", c)
	}
	var channel legend.Barrel[<-chan int, any]
	ch := make(chan int)
	if c := values.Fill(&channel, ch); !c.OK() || channel.Get() != ch || !channel.Present() {
		t.Fatal("channel", c)
	}
	var nilSlice []int
	if c := values.Fill(&slice, nilSlice); !c.OK() || !slice.Present() || slice.Get() != nil {
		t.Fatal("typed nil slice", c)
	}
}

func TestCheckedDestinationsAndPresence(t *testing.T) {
	var q legend.Qualifier
	var outer custom
	if q.Present() || outer.Present() {
		t.Fatal("zero presence")
	}
	if !values.Mark(&q).OK() || !q.Present() || outer.Present() {
		t.Fatal("bare mark")
	}
	if !values.Mark(&outer).OK() || !outer.Present() {
		t.Fatal("promoted writer")
	}
	var nilFeature *legend.Feature
	var nilQualifier *legend.Qualifier
	for _, d := range []any{nil, nilFeature, legend.Feature{}, new(int), new(carriers.Feature)} {
		if values.Fill(d, "x").OK() {
			t.Fatalf("accepted invalid destination %T", d)
		}
	}
	for _, d := range []any{nil, nilQualifier, legend.Qualifier{}, new(int), &broken{}} {
		if values.Mark(d).OK() {
			t.Fatalf("accepted invalid destination %T", d)
		}
	}
}

func TestPublicMarkersHaveNoSetters(t *testing.T) {
	for _, marker := range []any{&legend.Feature{}, &legend.Barrel[int, any]{}, &legend.Qualifier{}} {
		for _, name := range []string{"Set", "Fill", "Mark"} {
			if _, ok := reflect.TypeOf(marker).MethodByName(name); ok {
				t.Fatalf("%T exposes %s", marker, name)
			}
		}
	}
}
