package traversal

import (
	"reflect"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/values"
	"github.com/qlustra/sail/legend"
)

type walkLandmark struct{ legend.Landmark }

type walkLane struct {
	legend.Lane
	First  walkLandmark
	Second walkLandmark
}

type walkRoot struct {
	legend.Legend
	One walkLane
	Two walkLane
}

type nestedLandmark struct{ legend.Landmark }
type nestedRegion struct {
	legend.Region
	Lane nestedLaneChild
}
type nestedLaneChild struct {
	legend.Lane
	Landmark nestedLandmark
}
type multiLandmark []nestedLandmark

type nestedLane struct {
	legend.Lane
	Region nestedRegion
	Many   multiLandmark
}

type nestedRoot struct {
	legend.Legend
	Outer nestedLane
}

type embeddedChart struct{ Value int }
type recipeForWalk struct {
	embeddedChart
	Other int
}

func field(t *testing.T, typ reflect.Type, name string) *inspect.FieldDescriptor {
	t.Helper()
	f, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("missing field %s", name)
	}
	return inspect.CaptureFieldDescriptor(f)
}

func laneUse(t *testing.T, typ reflect.Type, name string, body *declarations.Lane) *declarations.LaneCarrier {
	t.Helper()
	property, check := capture.TypeReflection(typ).AsStruct()
	if !check.OK() {
		t.Fatal(check)
	}
	p, check := property.Property(name)
	if !check.OK() {
		t.Fatal(check)
	}
	return &declarations.LaneCarrier{Carrier: &define.Carrier[declarations.Lane]{Property: p, Declaration: body}}
}

func landmarkSpot(t *testing.T, typ reflect.Type, name string) *declarations.Spot {
	t.Helper()
	property, check := capture.TypeReflection(typ).AsStruct()
	if !check.OK() {
		t.Fatal(check)
	}
	p, check := property.Property(name)
	if !check.OK() {
		t.Fatal(check)
	}
	body := &declarations.Landmark{Family: name}
	return &declarations.Spot{For: declarations.SpotForLandmark, Kind: declarations.SingleSpot, Landmark: &define.Carrier[declarations.Landmark]{Property: p, Declaration: body}}
}

func property(t *testing.T, typ reflect.Type, name string) *capture.PropertyInterrogator {
	t.Helper()
	s, check := capture.TypeReflection(typ).AsStruct()
	if !check.OK() {
		t.Fatal(check)
	}
	p, check := s.Property(name)
	if !check.OK() {
		t.Fatal(check)
	}
	return p
}

func TestSurveyNestedMultiSpotChartOrderAndPrefixes(t *testing.T) {
	rootType, laneType, regionType := reflect.TypeFor[nestedRoot](), reflect.TypeFor[nestedLane](), reflect.TypeFor[nestedRegion]()
	regionBody := &declarations.Region{Family: "region"}
	childLaneBody := &declarations.Lane{}
	childSpot := &declarations.Spot{For: declarations.SpotForLandmark, Kind: declarations.SingleSpot,
		Landmark: &define.Carrier[declarations.Landmark]{Property: property(t, reflect.TypeFor[nestedLaneChild](), "Landmark"), Declaration: &declarations.Landmark{Family: "child"}}}
	if !childLaneBody.AddSpot(childSpot).OK() {
		t.Fatal("add nested lane spot")
	}
	childLane := &declarations.LaneCarrier{Carrier: &define.Carrier[declarations.Lane]{Property: property(t, regionType, "Lane"), Declaration: childLaneBody}}
	regionBody.Lanes = []*declarations.LaneCarrier{childLane}
	outerBody := &declarations.Lane{}
	regionSpot := &declarations.Spot{For: declarations.SpotForRegion, Kind: declarations.SingleSpot,
		Region: &define.Carrier[declarations.Region]{Property: property(t, laneType, "Region"), Declaration: regionBody}}
	multiSpot := &declarations.Spot{For: declarations.SpotForLandmark, Kind: declarations.MultiSpot,
		Landmark: &define.Carrier[declarations.Landmark]{Property: property(t, laneType, "Many"), Declaration: &declarations.Landmark{Family: "many"}}}
	if !outerBody.AddSpot(regionSpot).OK() || !outerBody.AddSpot(multiSpot).OK() {
		t.Fatal("add outer spots")
	}
	outerLane := &declarations.LaneCarrier{Carrier: &define.Carrier[declarations.Lane]{Property: property(t, rootType, "Outer"), Declaration: outerBody}}
	grammar := &declarations.Legend{Lanes: []*declarations.LaneCarrier{outerLane}}
	model := &nestedRoot{}
	model.Outer.Many = make(multiLandmark, 2)
	for _, item := range []struct {
		marker any
		index  []int
	}{
		{&model.Outer.Region, []int{0, 0}},  // embedding path; parent precedes its nested occurrence
		{&model.Outer.Many[0], []int{0, 2}}, // deliberately reverse slice and Chart order
		{&model.Outer.Many[1], []int{0, 1}},
		{&model.Outer.Region.Lane.Landmark, []int{0, 0, 0}},
	} {
		if check := values.RecordOccurrence(item.marker, capture.Type[recipeForWalk](), item.index); !check.OK() {
			t.Fatal(check)
		}
	}
	type event struct {
		depth              int
		token              any
		incoming, returned string
	}
	var got []event
	err := Survey(grammar, model, true, "root", func(depth int, _ *declarations.LaneCarrier, token any, inherited string) string {
		name := "region"
		switch token {
		case &model.Outer.Many[0]:
			name = "many-0"
		case &model.Outer.Many[1]:
			name = "many-1"
		case &model.Outer.Region.Lane.Landmark:
			name = "nested-landmark"
		}
		result := inherited + ">" + name
		got = append(got, event{depth, token, inherited, result})
		return result
	})
	if err != nil {
		t.Fatal(err)
	}
	wantTokens := []any{&model.Outer.Region, &model.Outer.Region.Lane.Landmark, &model.Outer.Many[1], &model.Outer.Many[0]}
	wantDepths := []int{1, 2, 1, 1}
	wantIncoming := []string{"root", "root>region", "root", "root"}
	wantReturned := []string{"root>region", "root>region>nested-landmark", "root>many-1", "root>many-0"}
	if len(got) != len(wantTokens) {
		t.Fatalf("events: %#v", got)
	}
	for i := range got {
		if got[i].token != wantTokens[i] || got[i].depth != wantDepths[i] || got[i].incoming != wantIncoming[i] || got[i].returned != wantReturned[i] {
			t.Fatalf("event %d = %#v; want token %p depth %d incoming %q returned %q", i, got[i], wantTokens[i], wantDepths[i], wantIncoming[i], wantReturned[i])
		}
	}
}

func TestSurveyDepthFirstAndSiblingPrefixes(t *testing.T) {
	rootType, laneType := reflect.TypeFor[walkRoot](), reflect.TypeFor[walkLane]()
	laneBody := &declarations.Lane{}
	first, second := landmarkSpot(t, laneType, "First"), landmarkSpot(t, laneType, "Second")
	if !laneBody.AddSpot(first).OK() || !laneBody.AddSpot(second).OK() {
		t.Fatal("could not assemble lane")
	}
	one := laneUse(t, rootType, "One", laneBody)
	two := laneUse(t, rootType, "Two", laneBody)
	one.ID, two.ID = "one", "two"
	grammar := &declarations.Legend{Lanes: []*declarations.LaneCarrier{one, two}}
	model := &walkRoot{}
	// Mark present occurrences using the same private occurrence writer used by population.
	for _, item := range []struct {
		marker *walkLandmark
		index  []int
	}{
		{&model.One.First, []int{0}}, {&model.One.Second, []int{1}},
		{&model.Two.First, []int{2}},
	} {
		if check := values.RecordOccurrence(item.marker, capture.Type[walkRoot](), item.index); !check.OK() {
			t.Fatal(check)
		}
	}
	type event struct {
		depth          int
		lane           *declarations.LaneCarrier
		token          any
		prefix, result string
	}
	var got []event
	if err := Survey(grammar, model, false, "root", func(depth int, lane *declarations.LaneCarrier, token any, prefix string) string {
		name := "first"
		if token == &model.One.Second {
			name = "second"
		}
		if lane == two {
			name = "two-" + name
		} else {
			name = "one-" + name
		}
		result := prefix + "/" + name
		got = append(got, event{depth, lane, token, prefix, result})
		return result
	}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].depth != 1 || got[1].depth != 1 || got[2].depth != 1 {
		t.Fatalf("walk events: %#v", got)
	}
	if got[0].lane != one || got[1].lane != one || got[2].lane != two {
		t.Fatal("incoming LaneCarrier identity/order not retained")
	}
	if got[0].token != &model.One.First || got[1].token != &model.One.Second || got[2].token != &model.Two.First {
		t.Fatal("visitor did not receive stored token identities")
	}
	for i, want := range []string{"root", "root", "root"} {
		if got[i].prefix != want {
			t.Fatalf("event %d inherited prefix %q, want %q", i, got[i].prefix, want)
		}
	}
	if got[0].result != "root/one-first" || got[1].result != "root/one-second" || got[2].result != "root/two-first" {
		t.Fatalf("sibling prefix results: %#v", got)
	}
}
