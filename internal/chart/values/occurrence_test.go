package values_test

import (
	"reflect"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/values"
	"github.com/qlustra/sail/legend"
)

type recipeRoot struct {
	Field  int
	Other  string
	Nested struct {
		Ignored bool
		Field   int
	}
}
type authoredRegion struct{ legend.Region }
type authoredLandmark struct{ legend.Landmark }
type brokenRegion struct{ *legend.Region }

func TestOccurrenceRecordSnapshots(t *testing.T) {
	var region authoredRegion
	if region.Present() {
		t.Fatal("fresh occurrence is present")
	}
	if loc, check := region.ChartLocation(); check.OK() || loc != nil {
		t.Fatal("fresh occurrence has a location")
	}
	root := capture.Type[recipeRoot]()
	index := []int{2, 1}
	if check := values.RecordOccurrence(&region, root, index); !check.OK() {
		t.Fatal(check)
	}
	index[0] = 99
	root.TypeInspector = capture.Type[int]().TypeInspector
	loc, check := region.ChartLocation()
	if !check.OK() || !region.Present() {
		t.Fatal(check)
	}
	assertLocation := func(loc *values.ChartLocation) {
		t.Helper()
		if loc.ChartType().CapturedType() != reflect.TypeFor[recipeRoot]() || !reflect.DeepEqual(loc.Index(), []int{2, 1}) {
			t.Fatal("recorded location changed")
		}
	}
	assertLocation(loc)
	returnedIndex := loc.Index()
	returnedIndex[0] = 88
	returnedType := loc.ChartType()
	returnedType.TypeInspector = capture.Type[int]().TypeInspector
	assertLocation(loc)
	*loc = values.ChartLocation{}
	loc, check = region.ChartLocation()
	if !check.OK() {
		t.Fatal(check)
	}
	assertLocation(loc)

	copyOfRegion := region
	var sibling authoredRegion
	var landmark authoredLandmark
	if sibling.Present() || landmark.Present() {
		t.Fatal("new cells share occurrence state")
	}
	if check := values.RecordOccurrence(&landmark, capture.Type[recipeRoot](), []int{0}); !check.OK() || !landmark.Present() {
		t.Fatal("promoted Landmark writer", check)
	}
	if check := values.RecordOccurrence(&region, capture.Type[recipeRoot](), []int{1}); !check.OK() {
		t.Fatal(check)
	}
	loc, check = copyOfRegion.ChartLocation()
	if !check.OK() {
		t.Fatal(check)
	}
	assertLocation(loc)
}

func TestOccurrenceCheckedWrites(t *testing.T) {
	root := capture.Type[recipeRoot]()
	var nilRegion *legend.Region
	for _, destination := range []any{nil, nilRegion, legend.Region{}, new(int), &brokenRegion{}, &legend.Feature{}, &legend.Qualifier{}} {
		if values.RecordOccurrence(destination, root, []int{0}).OK() {
			t.Fatalf("accepted destination %T", destination)
		}
	}
	var region legend.Region
	bad := []struct {
		root *capture.TypeInterrogator
		path []int
	}{
		{nil, []int{0}}, {&capture.TypeInterrogator{}, []int{0}},
		{capture.TypeReflection(nil), []int{0}}, {capture.Type[int](), []int{0}},
		{capture.Type[*recipeRoot](), []int{0}}, {root, nil}, {root, []int{}}, {root, []int{0, -1}},
	}
	for _, input := range bad {
		if values.RecordOccurrence(&region, input.root, input.path).OK() || region.Present() {
			t.Fatal("bad write created a record")
		}
	}
	if check := values.RecordOccurrence(&region, root, []int{0}); !check.OK() {
		t.Fatal(check)
	}
	for _, input := range bad {
		if values.RecordOccurrence(&region, input.root, input.path).OK() {
			t.Fatal("bad write accepted")
		}
		loc, check := region.ChartLocation()
		if !check.OK() || !reflect.DeepEqual(loc.Index(), []int{0}) || loc.ChartType().CapturedType() != reflect.TypeFor[recipeRoot]() {
			t.Fatal("bad write changed existing record")
		}
	}
}

func TestOccurrencePublicSurface(t *testing.T) {
	for _, marker := range []any{&legend.Region{}, &legend.Landmark{}, &authoredRegion{}, &authoredLandmark{}} {
		for _, name := range []string{"Set", "Fill", "Mark", "RecordOccurrence"} {
			if _, ok := reflect.TypeOf(marker).MethodByName(name); ok {
				t.Fatalf("%T exposes %s", marker, name)
			}
		}
	}
	for _, marker := range []any{&legend.Legend{}, &legend.Lane{}, &legend.Feature{}, &legend.Qualifier{}, &legend.Barrel[int, any]{}} {
		if _, ok := reflect.TypeOf(marker).MethodByName("ChartLocation"); ok {
			t.Fatalf("%T exposes ChartLocation", marker)
		}
	}
}
