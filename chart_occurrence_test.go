package sail_test

import (
	"reflect"
	"testing"

	"github.com/qlustra/intel/logic"
	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type occurrenceReader interface {
	Present() bool
	ChartLocation() (*sail.ChartLocation, logic.Check)
}

func occurrenceCheck(t *testing.T, item occurrenceReader, root reflect.Type, path []int, name string) {
	t.Helper()
	if !item.Present() {
		t.Fatal("occurrence is absent")
	}
	location, check := item.ChartLocation()
	if !check.OK() || location == nil {
		t.Fatalf("location unavailable: %v", check)
	}
	typ := location.ChartType()
	if typ == nil || typ.CapturedType() != root || !reflect.DeepEqual(location.Index(), path) {
		t.Fatalf("location type=%v path=%v, want %v %v", typ, location.Index(), root, path)
	}
	if field := root.FieldByIndex(path); field.Name != name {
		t.Fatalf("path %v reaches %s, want %s", path, field.Name, name)
	}
}

type occurrenceEmptyRegion struct {
	legend.Region `sail:"region" family:"region"`
}
type occurrenceEmptyLandmark struct {
	legend.Landmark `sail:"landmark" family:"leaf"`
}
type occurrenceSinglesLane struct {
	legend.Lane
	Region occurrenceEmptyRegion   `sail:"spot" spot:"single"`
	Leaf   occurrenceEmptyLandmark `sail:"spot" spot:"single"`
}
type occurrenceSinglesLegend struct {
	legend.Legend `sail:"legend" mark:"occ"`
	Items         occurrenceSinglesLane `sail:"lane"`
}
type occurrenceAbsentRecipe struct{ Ignored int }
type occurrenceSingleRecipe struct {
	Ignored int
	Region  struct{} `occ:"region"`
	Leaf    bool     `occ:"leaf"`
}

type occurrenceDescribedLandmark struct {
	legend.Landmark `sail:"landmark" family:"leaf"`
	Name            legend.Feature          `sail:"feature" symbol:"name"`
	Value           legend.Barrel[int, any] `sail:"barrel" populate:"Value"`
}
type occurrenceDescribedLane struct {
	legend.Lane
	Leaf occurrenceDescribedLandmark `sail:"spot" spot:"single"`
}
type occurrenceDescribedLegend struct {
	legend.Legend `sail:"legend" mark:"occ"`
	Items         occurrenceDescribedLane `sail:"lane"`
}

func TestOccurrenceWithoutDescriptiveValues(t *testing.T) {
	chart, err := sail.Compile[occurrenceDescribedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[occurrenceSingleRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	occurrenceCheck(t, m.Items.Leaf, reflect.TypeFor[occurrenceSingleRecipe](), []int{2}, "Leaf")
	if m.Items.Leaf.Name.Present() || m.Items.Leaf.Value.Present() {
		t.Fatal("occurrence presence must not imply Feature or Barrel availability")
	}
}

func TestOccurrenceOptionalAndEmpty(t *testing.T) {
	chart, err := sail.Compile[occurrenceSinglesLegend]()
	if err != nil {
		t.Fatal(err)
	}
	absent, err := chart.Build[occurrenceAbsentRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []occurrenceReader{absent.Items.Region, absent.Items.Leaf} {
		location, check := item.ChartLocation()
		if item.Present() || check.OK() || location != nil {
			t.Fatalf("absent occurrence: present=%v location=%v check=%v", item.Present(), location, check)
		}
	}
	matched, err := chart.Build[occurrenceSingleRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	root := reflect.TypeFor[occurrenceSingleRecipe]()
	occurrenceCheck(t, matched.Items.Region, root, []int{1}, "Region")
	occurrenceCheck(t, matched.Items.Leaf, root, []int{2}, "Leaf")
}

type occurrenceManyLane struct {
	legend.Lane
	Regions []occurrenceEmptyRegion   `sail:"spot" spot:"multi"`
	Leaves  []occurrenceEmptyLandmark `sail:"spot" spot:"multi"`
}
type occurrenceManyLegend struct {
	legend.Legend `sail:"legend" mark:"occ"`
	Items         occurrenceManyLane `sail:"lane"`
}
type occurrenceCarrier struct{ Live string }
type occurrenceManyRecipe struct {
	Ignored int
	A       occurrenceCarrier `occ:"region"`
	B       occurrenceCarrier `occ:"region"`
	C       occurrenceCarrier `occ:"region"`
	D       occurrenceCarrier `occ:"region"`
	E       occurrenceCarrier `occ:"region"`
	F       bool              `occ:"leaf"`
	G       bool              `occ:"leaf"`
	H       bool              `occ:"leaf"`
	I       bool              `occ:"leaf"`
	J       bool              `occ:"leaf"`
}
type occurrenceOtherRecipe occurrenceManyRecipe
type occurrenceAlias = occurrenceManyRecipe

func TestOccurrenceTypeAndIndependentSnapshots(t *testing.T) {
	chart, err := sail.Compile[occurrenceManyLegend]()
	if err != nil {
		t.Fatal(err)
	}
	a, err := chart.Build[occurrenceManyRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	again, err := chart.Build[occurrenceAlias]()
	if err != nil {
		t.Fatal(err)
	}
	other, err := chart.Build[occurrenceOtherRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	live := occurrenceManyRecipe{A: occurrenceCarrier{Live: "configured"}, B: occurrenceCarrier{Live: "different"}, F: true}
	populated, err := chart.Populate(live)
	if err != nil {
		t.Fatal(err)
	}
	verify := func(m *occurrenceManyLegend, root reflect.Type) {
		t.Helper()
		if len(m.Items.Regions) != 5 || len(m.Items.Leaves) != 5 {
			t.Fatalf("counts: %d %d", len(m.Items.Regions), len(m.Items.Leaves))
		}
		for i, name := range []string{"A", "B", "C", "D", "E"} {
			occurrenceCheck(t, m.Items.Regions[i], root, []int{i + 1}, name)
		}
		for i, name := range []string{"F", "G", "H", "I", "J"} {
			occurrenceCheck(t, m.Items.Leaves[i], root, []int{i + 6}, name)
		}
	}
	root := reflect.TypeFor[occurrenceManyRecipe]()
	verify(a, root)
	verify(again, root)
	verify(populated, root)
	verify(other, reflect.TypeFor[occurrenceOtherRecipe]())
	if reflect.TypeFor[occurrenceOtherRecipe]() == root {
		t.Fatal("defined root types unexpectedly identical")
	}
	loc, check := a.Items.Regions[0].ChartLocation()
	if !check.OK() {
		t.Fatal(check)
	}
	path := loc.Index()
	path[0] = 100
	wrapper := loc.ChartType()
	if wrapper == loc.ChartType() {
		t.Fatal("ChartType returned same wrapper")
	}
	*wrapper = *loc.ChartType()
	wrapper.TypeInspector = nil
	*loc = sail.ChartLocation{}
	verify(a, root)
	verify(again, root)
	verify(populated, root)
	verify(other, reflect.TypeFor[occurrenceOtherRecipe]())
	a.Items.Regions[0] = a.Items.Regions[1]
	occurrenceCheck(t, again.Items.Regions[0], root, []int{1}, "A")
	if live.A.Live != "configured" || live.B.Live != "different" || !live.F {
		t.Fatal("Populate modified live recipe")
	}
}

type occurrenceNestedLane struct {
	legend.Lane
	Leaves []occurrenceEmptyLandmark `sail:"spot" spot:"multi"`
}
type occurrenceNestedRegion struct {
	legend.Region `sail:"region" family:"region"`
	Children      occurrenceNestedLane `sail:"lane"`
}
type occurrenceNestedRootLane struct {
	legend.Lane
	Regions []occurrenceNestedRegion `sail:"spot" spot:"multi"`
}
type occurrenceNestedLegend struct {
	legend.Legend `sail:"legend" mark:"occ"`
	Roots         occurrenceNestedRootLane `sail:"lane"`
}
type OccurrenceEmbeddedLeaves struct {
	Skip int
	Leaf bool `occ:"leaf"`
}
type occurrenceNestedBody struct {
	Ignored int
	OccurrenceEmbeddedLeaves
	Last bool `occ:"leaf"`
}
type OccurrenceEmbeddedRoots struct {
	Ignored int
	First   occurrenceNestedBody `occ:"region"`
}
type occurrenceNestedRecipe struct {
	Ignored int
	OccurrenceEmbeddedRoots
	Second occurrenceNestedBody `occ:"region"`
}

func TestOccurrenceNestedPromotions(t *testing.T) {
	chart, err := sail.Compile[occurrenceNestedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[occurrenceNestedRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Roots.Regions) != 2 {
		t.Fatalf("regions=%d", len(m.Roots.Regions))
	}
	root := reflect.TypeFor[occurrenceNestedRecipe]()
	for i, prefix := range [][]int{{1, 1}, {2}} {
		r := m.Roots.Regions[i]
		name := []string{"First", "Second"}[i]
		occurrenceCheck(t, r, root, prefix, name)
		if len(r.Children.Leaves) != 2 {
			t.Fatalf("leaves=%d", len(r.Children.Leaves))
		}
		occurrenceCheck(t, r.Children.Leaves[0], root, append(append([]int{}, prefix...), 1, 1), "Leaf")
		occurrenceCheck(t, r.Children.Leaves[1], root, append(append([]int{}, prefix...), 2), "Last")
	}
}

func TestOccurrenceRecursiveRegion(t *testing.T) {
	chart, err := sail.Compile[recursiveLegend]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[recursiveRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Roots.Nodes) != 2 || len(m.Roots.Nodes[0].Children.Nodes) != 2 || len(m.Roots.Nodes[0].Children.Nodes[0].Children.Nodes) != 1 {
		t.Fatal("unexpected recursive tree")
	}
	root := reflect.TypeFor[recursiveRecipe]()
	first := m.Roots.Nodes[0]
	child := first.Children.Nodes[0]
	occurrenceCheck(t, first, root, []int{0}, "First")
	occurrenceCheck(t, m.Roots.Nodes[1], root, []int{1}, "Second")
	occurrenceCheck(t, child, root, []int{0, 1}, "First")
	occurrenceCheck(t, first.Children.Nodes[1], root, []int{0, 2}, "Second")
	occurrenceCheck(t, child.Children.Nodes[0], root, []int{0, 1, 2}, "Grandchild")
	occurrenceCheck(t, child.Flags.Flags[1], root, []int{0, 1, 1}, "Extra")
}
