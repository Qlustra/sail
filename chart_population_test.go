package sail_test

import (
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type populationCommand struct {
	legend.Region `sail:"region" family:"command"`
	Name          legend.Feature `sail:"feature" symbol:"name"`
	Help          legend.Feature `sail:"feature" symbol:"help"`
}

type populationLane struct {
	legend.Lane
	Command populationCommand `sail:"spot" spot:"single,required"`
}

type populationLegend struct {
	legend.Legend `sail:"legend" mark:"dome"`
	Commands      populationLane `sail:"lane"`
}

type populationAppCommand struct{ Live string }
type populationRecipe struct {
	Before *int
	Other  *int                 `dome:"other"`
	Main   populationAppCommand `dome:"command" name:"run" help:"Run it"`
	After  *int
	More   *int `dome:"other"`
}
type populationEmptyHelp struct {
	Main populationAppCommand `dome:"command" name:"empty" help:""`
}
type populationMissingHelp struct {
	Main populationAppCommand `dome:"command" name:"missing"`
}

func populationChart(t *testing.T) *sail.Navigator[populationLegend] {
	t.Helper()
	c, err := sail.Compile[populationLegend]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	return c
}

func checkPopulation(t *testing.T, m *populationLegend, name, help string, helpPresent bool) {
	t.Helper()
	if m == nil {
		t.Fatal("nil model")
	}
	c := &m.Commands.Command
	if !c.Name.Present() || c.Name.Get() != name || c.Help.Present() != helpPresent || c.Help.Get() != help {
		t.Fatalf("features: name=(%q,%v) help=(%q,%v), want (%q,true) (%q,%v)", c.Name.Get(), c.Name.Present(), c.Help.Get(), c.Help.Present(), name, help, helpPresent)
	}
}

func TestPopulationFactoryAndReuse(t *testing.T) {
	c := populationChart(t)
	built, err := c.Build[populationRecipe]()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	n := 42
	input := populationRecipe{Before: &n, Other: &n, Main: populationAppCommand{Live: "nonzero"}, After: &n, More: &n}
	populated, err := c.Populate(input)
	if err != nil {
		t.Fatalf("Populate: %v", err)
	}
	missing, err := c.Build[populationMissingHelp]()
	if err != nil {
		t.Fatalf("missing help: %v", err)
	}
	empty, err := c.Build[populationEmptyHelp]()
	if err != nil {
		t.Fatalf("empty help: %v", err)
	}
	again, err := c.Build[populationRecipe]()
	if err != nil {
		t.Fatalf("repeated Build: %v", err)
	}
	if built == populated || built == again || populated == again || missing == empty {
		t.Fatal("Chart reused a model")
	}
	checkPopulation(t, built, "run", "Run it", true)
	checkPopulation(t, populated, "run", "Run it", true)
	checkPopulation(t, missing, "missing", "", false)
	checkPopulation(t, empty, "empty", "", true)
	checkPopulation(t, again, "run", "Run it", true)
	if input.Main.Live != "nonzero" || n != 42 {
		t.Fatal("Populate mutated the live Recipe")
	}
	for range 8 {
		t.Run("concurrent reuse", func(t *testing.T) {
			t.Parallel()
			m, err := c.Build[populationRecipe]()
			if err != nil {
				t.Fatal(err)
			}
			checkPopulation(t, m, "run", "Run it", true)
		})
	}
}

type populationPointerCommand struct {
	Main *populationAppCommand `dome:"command" name:"pointer"`
}
type populationAbsent struct{ Other string }

func TestPopulationAdmissionAndRequiredSpot(t *testing.T) {
	c := populationChart(t)
	if m, err := c.Build[*populationRecipe](); err == nil || m != nil {
		t.Fatalf("pointer root Build: model=%v error=%v", m, err)
	}
	var root *populationRecipe
	if m, err := c.Populate(root); err == nil || m != nil {
		t.Fatalf("typed nil root: model=%v error=%v", m, err)
	}
	if m, err := c.Populate(&populationRecipe{}); err == nil || m != nil {
		t.Fatalf("pointer root Populate: model=%v error=%v", m, err)
	}
	if m, err := c.Build[populationPointerCommand](); err == nil || m != nil {
		t.Fatalf("pointer command Build: model=%v error=%v", m, err)
	}
	if m, err := c.Populate(populationPointerCommand{}); err == nil || m != nil {
		t.Fatalf("nil pointer command: model=%v error=%v", m, err)
	}
	if m, err := c.Populate(populationPointerCommand{Main: &populationAppCommand{}}); err == nil || m != nil {
		t.Fatalf("pointer command: model=%v error=%v", m, err)
	}
	if m, err := c.Populate(nil); err == nil || m != nil {
		t.Fatalf("untyped nil: model=%v error=%v", m, err)
	}
	if m, err := c.Build[populationAbsent](); err == nil || m != nil {
		t.Fatalf("missing required command: model=%v error=%v", m, err)
	}
}

type PopulationEmbedded struct {
	Main populationAppCommand `dome:"command" name:"promoted" help:"embedded"`
}
type populationPromoted struct{ PopulationEmbedded }
type PopulationAmbiguousA struct {
	Main populationAppCommand `dome:"command" name:"a"`
}
type PopulationAmbiguousB struct {
	Main populationAppCommand `dome:"command" name:"b"`
}
type populationAmbiguous struct {
	PopulationAmbiguousA
	PopulationAmbiguousB
}
type populationHidden struct {
	PopulationEmbedded
	Main string
}
type populationShadow struct {
	PopulationEmbedded
	Main populationAppCommand `dome:"command" name:"direct" help:"shadow"`
}
type populationPointerEmbedding struct{ *PopulationEmbedded }
type PopulationUnrelated struct{ Value string }
type populationUnrelatedPointer struct {
	*PopulationUnrelated
	Main populationAppCommand `dome:"command" name:"safe" help:"unrelated"`
}

func TestPopulationRecipePromotion(t *testing.T) {
	c := populationChart(t)
	promoted, err := c.Build[populationPromoted]()
	if err != nil {
		t.Fatalf("promoted: %v", err)
	}
	checkPopulation(t, promoted, "promoted", "embedded", true)
	shadow, err := c.Build[populationShadow]()
	if err != nil {
		t.Fatalf("shadow: %v", err)
	}
	checkPopulation(t, shadow, "direct", "shadow", true)
	unrelated, err := c.Build[populationUnrelatedPointer]()
	if err != nil {
		t.Fatalf("unrelated pointer embedding: %v", err)
	}
	checkPopulation(t, unrelated, "safe", "unrelated", true)
	if m, err := c.Build[populationAmbiguous](); err == nil || m != nil {
		t.Fatalf("ambiguous Main: model=%v error=%v", m, err)
	}
	if m, err := c.Build[populationHidden](); err == nil || m != nil {
		t.Fatalf("hidden Main: model=%v error=%v", m, err)
	}
	if m, err := c.Build[populationPointerEmbedding](); err == nil || m != nil {
		t.Fatalf("pointer path to Main: model=%v error=%v", m, err)
	}
}

type PopulationFeatures struct {
	Name legend.Feature `sail:"feature" symbol:"name"`
	Help legend.Feature `sail:"feature" symbol:"help"`
}
type populationPromotedCommand struct {
	legend.Region `sail:"region" family:"command"`
	PopulationFeatures
}
type PopulationSpots struct {
	Command populationPromotedCommand `sail:"spot" spot:"single,required"`
}
type populationPromotedLane struct {
	legend.Lane
	PopulationSpots
}
type PopulationLanes struct {
	Commands populationPromotedLane `sail:"lane"`
}
type populationPromotedLegend struct {
	legend.Legend `sail:"legend" mark:"dome"`
	PopulationLanes
}

func TestPopulationPromotedDestinations(t *testing.T) {
	c, err := sail.Compile[populationPromotedLegend]()
	if err != nil || c == nil {
		t.Fatalf("Compile promoted destinations: chart=%v error=%v", c, err)
	}
	m, err := c.Build[populationRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: model=%v error=%v", m, err)
	}
	f := m.Commands.Command.PopulationFeatures
	if !f.Name.Present() || f.Name.Get() != "run" || !f.Help.Present() || f.Help.Get() != "Run it" {
		t.Fatalf("promoted features: name=(%q,%v) help=(%q,%v)", f.Name.Get(), f.Name.Present(), f.Help.Get(), f.Help.Present())
	}
}
