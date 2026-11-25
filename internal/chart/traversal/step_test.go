package traversal

import (
	"slices"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/exploration"
	"github.com/qlustra/sail/internal/chart/populating"
	"github.com/qlustra/sail/legend"
)

type command struct {
	legend.Region `sail:"region" family:"command"`
	Name          legend.Feature `sail:"feature" symbol:"name"`
	Commands      commands       `sail:"lane"`
}

type commands struct {
	legend.Lane
	Commands []command `sail:"spot" spot:"multi"`
}

type clientsLegend struct {
	legend.Legend `sail:"legend" mark:"cli"`
	Commands      commands `sail:"lane"`
}

func TestImmediateMatchesResumesAtSibling(t *testing.T) {
	type clients struct {
		A struct {
			Status struct{} `cli:"command" name:"status"`
		} `cli:"command" name:"client"`
		B struct {
			Version struct{} `cli:"command" name:"version"`
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
	program, err := (&populating.Populator[clientsLegend]{Grammar: grammar}).Populate(capture.Type[clients]())
	if err != nil {
		t.Fatal(err)
	}
	start, err := Start(grammar, program)
	if err != nil {
		t.Fatal(err)
	}
	var tested []*command
	named := func(name string) func(int, *declarations.LaneCarrier, any) (*command, bool) {
		return func(_ int, _ *declarations.LaneCarrier, pointer any) (*command, bool) {
			cmd := pointer.(*command)
			tested = append(tested, cmd)
			return cmd, cmd.Name.Get() == name
		}
	}
	clientsAtStart, err := start.immediateMatches(named("client"), false)
	if err != nil {
		t.Fatal(err)
	}
	a, b := &program.Commands.Commands[0], &program.Commands.Commands[1]
	status, version := &a.Commands.Commands[0], &b.Commands.Commands[0]
	var routes [][2]*command
	for position, client := range clientsAtStart {
		versions, err := position.immediateMatches(named("version"), false)
		if err != nil {
			t.Fatal(err)
		}
		for _, destination := range versions {
			routes = append(routes, [2]*command{client, destination})
		}
	}
	if !slices.Equal(tested, []*command{a, status, b, version}) {
		t.Fatal("expected to test A, then its status, then B, then its version")
	}
	if !slices.Equal(routes, [][2]*command{{b, version}}) {
		t.Fatal("only B → version should complete")
	}
	tested = nil
	for range clientsAtStart {
		break
	}
	if !slices.Equal(tested, []*command{a}) {
		t.Fatal("breaking after A must not run the predicate on B")
	}
}
