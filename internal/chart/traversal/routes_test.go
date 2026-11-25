package traversal

import (
	"slices"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/exploration"
	"github.com/qlustra/sail/internal/chart/populating"
	"github.com/qlustra/sail/internal/chart/values"
)

func TestRoutesDepthFirstOrderSnapshotsAndStop(t *testing.T) {
	type recipe struct {
		A struct {
			One struct{} `cli:"command" name:"one"`
			Two struct{} `cli:"command" name:"two"`
		} `cli:"command" name:"client"`
		B struct {
			One struct{} `cli:"command" name:"one"`
			Two struct{} `cli:"command" name:"two"`
		} `cli:"command" name:"client"`
	}
	explorer, err := exploration.New()
	if err != nil {
		t.Fatal(err)
	}
	grammar, err := explorer.Explore[clientsLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := (&populating.Populator[clientsLegend]{Grammar: grammar}).Populate(capture.Type[recipe]())
	if err != nil {
		t.Fatal(err)
	}
	start, err := Start(grammar, model)
	if err != nil {
		t.Fatal(err)
	}
	a, b := &model.Commands.Commands[0], &model.Commands.Commands[1]
	// Reverse Chart order independently at both levels without moving storage.
	for _, item := range []struct {
		cmd  *command
		path []int
	}{
		{a, []int{1}}, {b, []int{0}},
		{&a.Commands.Commands[0], []int{1, 1}}, {&a.Commands.Commands[1], []int{1, 0}},
		{&b.Commands.Commands[0], []int{0, 1}}, {&b.Commands.Commands[1], []int{0, 0}},
	} {
		if check := values.RecordOccurrence(item.cmd, capture.Type[recipe](), item.path); !check.OK() {
			t.Fatal(check)
		}
	}
	var tested []any
	match := func(depth int, lane *declarations.LaneCarrier, pointer any) (any, bool) {
		if lane == nil || depth < 1 || depth > 2 {
			t.Fatalf("invalid position: %d %v", depth, lane)
		}
		tested = append(tested, pointer)
		return pointer, true
	}
	steps := []WaypointMatch{match, match}
	for _, tc := range []struct {
		recipe bool
		want   [][]any
	}{
		{false, [][]any{{a, &a.Commands.Commands[0]}, {a, &a.Commands.Commands[1]}, {b, &b.Commands.Commands[0]}, {b, &b.Commands.Commands[1]}}},
		{true, [][]any{{b, &b.Commands.Commands[1]}, {b, &b.Commands.Commands[0]}, {a, &a.Commands.Commands[1]}, {a, &a.Commands.Commands[0]}}},
	} {
		var routes [][]any
		tested = nil
		if err := start.Routes(steps, tc.recipe, func(route []any) bool { routes = append(routes, route); return true }); err != nil {
			t.Fatal(err)
		}
		if len(routes) != len(tc.want) {
			t.Fatalf("routes: %v", routes)
		}
		for i := range routes {
			if !slices.Equal(routes[i], tc.want[i]) {
				t.Fatalf("route %d = %v, want %v", i, routes[i], tc.want[i])
			}
		}
		if len(tested) != 6 || tested[0] != tc.want[0][0] || tested[1] != tc.want[0][1] || tested[3] != tc.want[2][0] {
			t.Fatalf("not depth first: %v", tested)
		}
		routes[0][0] = nil
		if routes[1][0] != tc.want[1][0] {
			t.Fatal("route slices share storage")
		}
		tested = nil
		if err := start.Routes(steps, tc.recipe, func([]any) bool { return false }); err != nil {
			t.Fatal(err)
		}
		if len(tested) != 2 {
			t.Fatalf("stop tested %d candidates", len(tested))
		}
	}
	// A failed first branch must not prevent a later complete route.
	var routes [][]any
	second := func(_ int, _ *declarations.LaneCarrier, p any) (any, bool) { return p, p == &b.Commands.Commands[1] }
	if err := start.Routes([]WaypointMatch{match, second}, false, func(r []any) bool { routes = append(routes, r); return true }); err != nil {
		t.Fatal(err)
	}
	if len(routes) != 1 || !slices.Equal(routes[0], []any{b, &b.Commands.Commands[1]}) {
		t.Fatalf("routes: %v", routes)
	}
	count := 0
	if err := start.Routes(nil, false, func(r []any) bool {
		count++
		if len(r) != 0 {
			t.Fatal(r)
		}
		return true
	}); err != nil || count != 1 {
		t.Fatalf("empty route: %d %v", count, err)
	}
}

func TestRoutesPropagatesNestedError(t *testing.T) {
	// A valid first level whose selected Region has an invalid Lane descriptor.
	type recipe struct {
		A struct{} `cli:"command"`
		B struct{} `cli:"command"`
	}
	explorer, err := exploration.New()
	if err != nil {
		t.Fatal(err)
	}
	grammar, err := explorer.Explore[clientsLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := (&populating.Populator[clientsLegend]{Grammar: grammar}).Populate(capture.Type[recipe]())
	if err != nil {
		t.Fatal(err)
	}
	start, err := Start(grammar, model)
	if err != nil {
		t.Fatal(err)
	}
	// Substitute a Lane from the Legend, whose field does not exist in command.
	region := grammar.Lanes[0].Declaration.Spots()[0].Region.Declaration
	region.Lanes = []*declarations.LaneCarrier{grammar.Lanes[0]}
	visits := 0
	match := func(_ int, _ *declarations.LaneCarrier, p any) (any, bool) { visits++; return p, true }
	err = start.Routes([]WaypointMatch{match, match}, false, func([]any) bool { t.Fatal("unexpected route"); return true })
	if err == nil || visits != 1 {
		t.Fatalf("error = %v, visits = %d", err, visits)
	}
}
