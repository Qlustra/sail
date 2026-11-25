package sail_test

import (
	"errors"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type routeLegend struct {
	legend.Legend `sail:"legend" mark:"route" legend:"routes"`
	Commands      routeLane `sail:"lane"`
}

type routeLane struct {
	legend.Lane
	Commands []routeCommand `sail:"spot" spot:"multi"`
}

type routeCommand struct {
	legend.Region `sail:"region" family:"command"`
	Name          legend.Feature `sail:"feature" symbol:"name"`
	Commands      routeLane      `sail:"lane"`
}

type routeOneChart struct {
	Dead struct{} `route:"command" name:"client"`
	Live struct {
		Version struct{} `route:"command" name:"version"`
	} `route:"command" name:"client"`
}

type routeTwoChart struct {
	First struct {
		Version struct{} `route:"command" name:"version"`
	} `route:"command" name:"client"`
	Second struct {
		Version struct{} `route:"command" name:"version"`
	} `route:"command" name:"client"`
}

func TestNextAndCourseCompleteRouteCardinality(t *testing.T) {
	nav, err := sail.Compile[routeLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := nav.Build[routeOneChart]()
	if err != nil {
		t.Fatal(err)
	}
	client := sail.Select[*routeCommand](func(c *routeCommand) bool { return c.Name.Get() == "client" })
	version := sail.Select[*routeCommand](func(c *routeCommand) bool { return c.Name.Get() == "version" })
	selection, marker, err := nav.Next(model, client)
	if err != nil || !selection.Ambiguous() || selection.Matched() || marker != nil {
		t.Fatalf("immediate clients: matched=%v ambiguous=%v marker=%v err=%v", selection.Matched(), selection.Ambiguous(), marker, err)
	}
	missing, marker, err := nav.Next(model, version)
	if err != nil || missing.Matched() || missing.Ambiguous() || marker != nil {
		t.Fatalf("immediate version: matched=%v ambiguous=%v marker=%v err=%v", missing.Matched(), missing.Ambiguous(), marker, err)
	}
	live := &model.Commands.Commands[1]
	unique, marker, err := nav.Next(model, sail.Select[*routeCommand](func(c *routeCommand) bool { return c == live }))
	if err != nil || !unique.Matched() || unique.Ambiguous() || marker != live || unique.Marker() != live {
		t.Fatalf("unique client: marker=%p want=%p err=%v", marker, live, err)
	}
	child, marker, err := unique.Next(version)
	wantChild := &live.Commands.Commands[0]
	if err != nil || !child.Matched() || child.Ambiguous() || marker != wantChild || child.Marker() != wantChild {
		t.Fatalf("unique version: marker=%p want=%p err=%v", marker, wantChild, err)
	}

	plot := nav.PlotCourse()
	plot.Waypoint(client)
	plot.Waypoint(version)
	course, err := plot.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	for _, order := range []sail.TraversalOrder{sail.ChartOrder, sail.LegendOrder} {
		t.Run(string(order), func(t *testing.T) {
			check := func(r sail.Route) {
				t.Helper()
				if len(r) != 2 || r[0] != live || r[1] != wantChild {
					t.Fatalf("route = %v, want stored live client and version", r)
				}
			}
			first, err := course.FirstRoute(model, order)
			if err != nil {
				t.Fatal(err)
			}
			check(first)
			one, err := course.RequireOneRoute(model, order)
			if err != nil {
				t.Fatal(err)
			}
			check(one)
			all, err := course.AllRoutes(model, order)
			if err != nil || len(all) != 1 {
				t.Fatalf("all routes = %v, err=%v", all, err)
			}
			check(all[0])
		})
	}

	noPlot := nav.PlotCourse()
	noPlot.Waypoint(client)
	noPlot.Waypoint(sail.Select[*routeCommand](func(c *routeCommand) bool { return c.Name.Get() == "missing" }))
	noCourse, err := noPlot.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if r, err := noCourse.FirstRoute(model, sail.ChartOrder); !errors.Is(err, sail.ErrNoRoute) || r != nil {
		t.Fatalf("first missing = %v, %v", r, err)
	}
	if r, err := noCourse.RequireOneRoute(model, sail.ChartOrder); !errors.Is(err, sail.ErrNoRoute) || r != nil {
		t.Fatalf("unique missing = %v, %v", r, err)
	}
	if r, err := noCourse.AllRoutes(model, sail.ChartOrder); err != nil || len(r) != 0 {
		t.Fatalf("all missing = %v, %v", r, err)
	}

	two, err := nav.Build[routeTwoChart]()
	if err != nil {
		t.Fatal(err)
	}
	if r, err := course.RequireOneRoute(two, sail.ChartOrder); !errors.Is(err, sail.ErrAmbiguousRoute) || r != nil {
		t.Fatalf("ambiguous = %v, %v", r, err)
	}
	routes, err := course.AllRoutes(two, sail.ChartOrder)
	if err != nil || len(routes) != 2 {
		t.Fatalf("all = %v, %v", routes, err)
	}
	for i, r := range routes {
		parent := &two.Commands.Commands[i]
		if len(r) != 2 || r[0] != parent || r[1] != &parent.Commands.Commands[0] {
			t.Fatalf("route %d = %v", i, r)
		}
	}
	first, err := course.FirstRoute(two, sail.ChartOrder)
	if err != nil || len(first) != 2 || first[0] != routes[0][0] || first[1] != routes[0][1] {
		t.Fatalf("first = %v, %v", first, err)
	}
	routes[0][0] = nil
	if routes[1][0] != &two.Commands.Commands[1] || first[0] != &two.Commands.Commands[0] {
		t.Fatal("route slices share storage")
	}
}
