package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type multiCommands struct {
	legend.Lane
	Commands []populationCommand `sail:"spot" spot:"multi"`
}
type multiLegend struct {
	legend.Legend `sail:"legend" mark:"dome"`
	Commands      multiCommands `sail:"lane"`
}
type MultiRecipeEmbedded struct {
	Middle populationAppCommand `dome:"command" name:"middle" help:""`
}
type multiRecipe struct {
	First populationAppCommand `dome:"command" name:"first" help:"first help"`
	MultiRecipeEmbedded
	Last populationAppCommand `dome:"command" name:"last"`
}

func checkMultiCommands(t *testing.T, commands []populationCommand) {
	t.Helper()
	if len(commands) != 3 {
		t.Fatalf("commands count = %d, want 3", len(commands))
	}
	for i, want := range []struct {
		name, help string
		present    bool
	}{{"first", "first help", true}, {"middle", "", true}, {"last", "", false}} {
		got := commands[i]
		if !got.Name.Present() || got.Name.Get() != want.name || got.Help.Present() != want.present || got.Help.Get() != want.help {
			t.Fatalf("command %d: name=(%q,%v), help=(%q,%v)", i, got.Name.Get(), got.Name.Present(), got.Help.Get(), got.Help.Present())
		}
	}
}

func TestMultiPopulationRegions(t *testing.T) {
	chart, err := sail.Compile[multiLegend]()
	if err != nil {
		t.Fatal(err)
	}
	built, err := chart.Build[multiRecipe]()
	if err != nil || built == nil {
		t.Fatalf("Build: %v", err)
	}
	recipe := multiRecipe{First: populationAppCommand{Live: "live first"}, Last: populationAppCommand{Live: "live last"}}
	populated, err := chart.Populate(recipe)
	if err != nil || populated == nil {
		t.Fatalf("Populate: %v", err)
	}
	again, err := chart.Build[multiRecipe]()
	if err != nil || again == nil {
		t.Fatalf("again: %v", err)
	}
	checkMultiCommands(t, built.Commands.Commands)
	checkMultiCommands(t, populated.Commands.Commands)
	checkMultiCommands(t, again.Commands.Commands)
	built.Commands.Commands[0] = built.Commands.Commands[2]
	if built.Commands.Commands[1].Name.Get() != "middle" {
		t.Fatal("captures share storage")
	}
	checkMultiCommands(t, populated.Commands.Commands)
	checkMultiCommands(t, again.Commands.Commands)
	if recipe.First.Live != "live first" || recipe.Last.Live != "live last" {
		t.Fatal("live Recipe changed")
	}
	for range 4 {
		t.Run("concurrent", func(t *testing.T) {
			t.Parallel()
			m, err := chart.Build[multiRecipe]()
			if err != nil || m == nil {
				t.Fatalf("Build: %v", err)
			}
			checkMultiCommands(t, m.Commands.Commands)
		})
	}
}

type multiRequiredCommands struct {
	legend.Lane
	Commands []populationCommand `sail:"spot" spot:"multi,required"`
}
type multiRequiredLegend struct {
	legend.Legend `sail:"legend" mark:"dome"`
	Commands      multiRequiredCommands `sail:"lane"`
}
type multiOneRecipe struct {
	One populationAppCommand `dome:"command" name:"one"`
}
type multiBoundary struct {
	legend.Landmark `sail:"landmark" family:"boundary"`
}
type multiBoundaryLane struct {
	legend.Lane
	Boundary multiBoundary `sail:"spot" spot:"single"`
}
type multiBoundaryLegend struct {
	legend.Legend `sail:"legend" mark:"dome"`
	Commands      multiRequiredCommands `sail:"lane"`
	Boundary      multiBoundaryLane     `sail:"lane"`
}
type multiBoundaryRecipe struct {
	Boundary bool `dome:"boundary"`
}

func TestMultiPopulationMinimum(t *testing.T) {
	optional, err := sail.Compile[multiLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if m, err := optional.Build[populationAbsent](); err != nil || m == nil || len(m.Commands.Commands) != 0 {
		t.Fatalf("optional zero: model=%v error=%v", m, err)
	}
	required, err := sail.Compile[multiRequiredLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if m, err := required.Build[populationAbsent](); err == nil || m != nil {
		t.Fatalf("required zero: model=%v error=%v", m, err)
	}
	if m, err := required.Build[multiOneRecipe](); err != nil || m == nil || len(m.Commands.Commands) != 1 {
		t.Fatalf("required one: model=%v error=%v", m, err)
	}
	if m, err := required.Build[multiRecipe](); err != nil || m == nil || len(m.Commands.Commands) != 3 {
		t.Fatalf("required many: model=%v error=%v", m, err)
	}
	boundary, err := sail.Compile[multiBoundaryLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if m, err := boundary.Build[multiBoundaryRecipe](); err == nil || m != nil {
		t.Fatalf("required zero at boundary: model=%v error=%v", m, err)
	}
}

type multiBoolLane struct {
	legend.Lane
	Flags []orderBool `sail:"spot" spot:"multi"`
}
type multiBoolLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Flags         multiBoolLane     `sail:"lane" exploration:"eager"`
	Fallback      orderFallbackLane `sail:"lane"`
}
type multiBoolRecipe struct {
	First  bool `order:"flag" name:"first"`
	Second bool `order:"flag" name:"second"`
}
type multiBadBoolRecipe struct {
	First  bool   `order:"flag" name:"first"`
	Second bool   `order:"flag" name:"second"`
	Bad    string `order:"flag"`
}
type multiPointerBoolRecipe struct {
	Bad *bool `order:"flag"`
}
type multiSliceBoolRecipe struct {
	Bad []bool `order:"flag"`
}

func TestMultiPopulationAcceptance(t *testing.T) {
	chart, err := sail.Compile[multiBoolLegend]()
	if err != nil {
		t.Fatal(err)
	}
	first, err := chart.Build[multiBoolRecipe]()
	if err != nil || first == nil {
		t.Fatalf("valid: %v", err)
	}
	if len(first.Flags.Flags) != 2 || first.Flags.Flags[0].Name.Get() != "first" || first.Flags.Flags[1].Name.Get() != "second" {
		t.Fatalf("flags: %+v", first.Flags.Flags)
	}
	if m, err := chart.Build[multiBadBoolRecipe](); err == nil || m != nil || !strings.Contains(err.Error(), "acceptance") {
		t.Fatalf("later bad flag: model=%v error=%v", m, err)
	}
	if m, err := chart.Build[multiPointerBoolRecipe](); err == nil || m != nil {
		t.Fatalf("pointer: model=%v error=%v", m, err)
	}
	if m, err := chart.Build[multiSliceBoolRecipe](); err == nil || m != nil {
		t.Fatalf("slice: model=%v error=%v", m, err)
	}
	again, err := chart.Build[multiBoolRecipe]()
	if err != nil || again == nil {
		t.Fatalf("after failure: %v", err)
	}
	again.Flags.Flags[0] = again.Flags.Flags[1]
	if first.Flags.Flags[0].Name.Get() != "first" {
		t.Fatal("failure or reuse aliased previous Legend")
	}
}

type multiNestedCommand struct {
	legend.Region `sail:"region" family:"command"`
	Flags         multiBoolLane `sail:"lane"`
}
type multiNestedCommandLane struct {
	legend.Lane
	Command multiNestedCommand `sail:"spot" spot:"single,required"`
}
type multiPartitionLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Global        multiBoolLane          `sail:"lane" exploration:"eager"`
	Command       multiNestedCommandLane `sail:"lane"`
	Local         multiBoolLane          `sail:"lane"`
}
type multiInnerRecipe struct {
	A bool `order:"flag" name:"nested-a"`
	B bool `order:"flag" name:"nested-b"`
}
type multiPartitionRecipe struct {
	G1      bool             `order:"flag" name:"global-a"`
	Unknown int              `order:"other"`
	G2      bool             `order:"flag" name:"global-b"`
	Command multiInnerRecipe `order:"command"`
	Other   string
	L1      bool `order:"flag" name:"local-a"`
	L2      bool `order:"flag" name:"local-b"`
	L3      bool `order:"flag" name:"local-c"`
}
type multiNoCommandRecipe struct {
	G bool `order:"flag"`
}

func checkMultiFlagNames(t *testing.T, flags []orderBool, names ...string) {
	t.Helper()
	if len(flags) != len(names) {
		t.Fatalf("flags count=%d want=%d", len(flags), len(names))
	}
	for i, name := range names {
		if !flags[i].Name.Present() || flags[i].Name.Get() != name {
			t.Fatalf("flag %d name=%q want=%q", i, flags[i].Name.Get(), name)
		}
	}
}
func TestMultiPopulationPartition(t *testing.T) {
	chart, err := sail.Compile[multiPartitionLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if m, err := chart.Build[multiNoCommandRecipe](); err == nil || m != nil {
		t.Fatalf("missing command: model=%v error=%v", m, err)
	}
	m, err := chart.Build[multiPartitionRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: %v", err)
	}
	checkMultiFlagNames(t, m.Global.Flags, "global-a", "global-b")
	checkMultiFlagNames(t, m.Command.Command.Flags.Flags, "nested-a", "nested-b")
	checkMultiFlagNames(t, m.Local.Flags, "local-a", "local-b", "local-c")
	m.Global.Flags[0] = m.Global.Flags[1]
	checkMultiFlagNames(t, m.Command.Command.Flags.Flags, "nested-a", "nested-b")
	checkMultiFlagNames(t, m.Local.Flags, "local-a", "local-b", "local-c")
}

type MultiPromotedSpots struct {
	Commands []populationCommand `sail:"spot" spot:"multi"`
}
type multiPromotedLane struct {
	legend.Lane
	MultiPromotedSpots
}
type multiPromotedLegend struct {
	legend.Legend `sail:"legend" mark:"dome"`
	Commands      multiPromotedLane `sail:"lane"`
}

type multiRequiredBoolLane struct {
	legend.Lane
	Flags []orderBool `sail:"spot" spot:"multi,required"`
}
type multiExtra struct {
	legend.Landmark `sail:"landmark" family:"extra"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}
type multiExtraLane struct {
	legend.Lane
	Extras []multiExtra `sail:"spot" spot:"multi"`
}
type multiGroup struct {
	legend.Region `sail:"region" family:"group"`
	Required      multiRequiredBoolLane `sail:"lane"`
	Optional      multiExtraLane        `sail:"lane"`
}
type multiGroupLane struct {
	legend.Lane
	Groups []multiGroup `sail:"spot" spot:"multi"`
}
type multiGroupsLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Groups        multiGroupLane `sail:"lane"`
}
type multiGroupInner struct {
	Flag  bool `order:"flag" name:"inner"`
	Extra bool `order:"extra" name:"extra"`
}
type multiGroupEmptyExtra struct {
	Flag bool `order:"flag" name:"other"`
}
type multiGroupMissing struct{ Other bool }
type multiGroupsRecipe struct {
	A multiGroupInner      `order:"group"`
	B multiGroupEmptyExtra `order:"group"`
	C multiGroupInner      `order:"group"`
}
type multiGroupsBadRecipe struct {
	A multiGroupInner   `order:"group"`
	B multiGroupMissing `order:"group"`
}

func TestMultiPopulationPromotedDestination(t *testing.T) {
	chart, err := sail.Compile[multiPromotedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := chart.Build[multiRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: %v", err)
	}
	checkMultiCommands(t, m.Commands.Commands)
}

func TestMultiPopulationNestedRegions(t *testing.T) {
	chart, err := sail.Compile[multiGroupsLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if m, err := chart.Build[multiGroupsBadRecipe](); err == nil || m != nil {
		t.Fatalf("missing nested required: model=%v error=%v", m, err)
	}
	m, err := chart.Build[multiGroupsRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: %v", err)
	}
	groups := m.Groups.Groups
	if len(groups) != 3 {
		t.Fatalf("groups count=%d", len(groups))
	}
	checkMultiFlagNames(t, groups[0].Required.Flags, "inner")
	checkMultiFlagNames(t, groups[1].Required.Flags, "other")
	checkMultiFlagNames(t, groups[2].Required.Flags, "inner")
	if len(groups[0].Optional.Extras) != 1 || len(groups[1].Optional.Extras) != 0 || len(groups[2].Optional.Extras) != 1 {
		t.Fatal("nested optional counts")
	}
	groups[0].Required.Flags[0] = groups[1].Required.Flags[0]
	checkMultiFlagNames(t, groups[2].Required.Flags, "inner")
}

type multiMixedLane struct {
	legend.Lane
	Commands []populationCommand `sail:"spot" spot:"multi,required"`
	Flags    []orderBool         `sail:"spot" spot:"multi,required"`
}

type multiMixedLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Members       multiMixedLane `sail:"lane"`
}

type multiMixedRecipe struct {
	FlagA    bool                 `order:"flag" name:"flag-a"`
	CommandA populationAppCommand `order:"command" name:"command-a"`
	FlagB    bool                 `order:"flag" name:"flag-b"`
	CommandB populationAppCommand `order:"command" name:"command-b"`
}

func TestMultiPopulationSpotMembersRemainUnordered(t *testing.T) {
	chart, err := sail.Compile[multiMixedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	legend, err := chart.Build[multiMixedRecipe]()
	if err != nil || legend == nil {
		t.Fatalf("mixed Spot captures: legend=%v error=%v", legend, err)
	}
	checkMultiFlagNames(t, legend.Members.Flags, "flag-a", "flag-b")
	commands := legend.Members.Commands
	if len(commands) != 2 || commands[0].Name.Get() != "command-a" || commands[1].Name.Get() != "command-b" {
		t.Fatal("interleaved families did not retain their independent occurrence order")
	}
}
