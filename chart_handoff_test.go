package sail

import (
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/exploration"
	"github.com/qlustra/sail/legend"
)

type handoffLeaf struct {
	legend.Landmark `sail:"landmark" family:"flag"`
	Name            legend.Feature   `sail:"feature" symbol:"name"`
	Enabled         legend.Qualifier `sail:"qualifier" symbol:"enabled"`
}
type handoffLane struct {
	legend.Lane
	Flag handoffLeaf `sail:"spot" spot:"single"`
}
type handoffModel struct {
	legend.Legend `sail:"legend" mark:"test"`
	First         handoffLane `sail:"lane" exploration:"eager"`
	Second        handoffLane `sail:"lane"`
}
type badHandoffLanes struct {
	legend.Legend `sail:"legend" mark:"test"`
	First         handoffLane `sail:"lane"`
	Second        handoffLane `sail:"lane"`
}
type collisionLeaf struct {
	legend.Landmark `sail:"landmark" family:"flag" discriminate:"enabled"`
	Enabled         legend.Qualifier `sail:"qualifier" symbol:"enabled"`
}
type collisionLane struct {
	legend.Lane
	Flag collisionLeaf `sail:"spot" spot:"single"`
}
type collisionModel struct {
	legend.Legend `sail:"legend" mark:"test"`
	Lane          collisionLane `sail:"lane"`
}

func TestCompiledGraphHandoff(t *testing.T) {
	c, err := Compile[handoffModel]()
	if err != nil || c == nil || c.grammar == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	g := c.grammar
	if len(g.Lanes) != 2 || g.Lanes[0].Declaration != g.Lanes[1].Declaration {
		t.Fatal("expected two uses of the same compiled lane")
	}
	if check := declarations.ValidateLaneExploration(g); !check.OK() {
		t.Fatal(check.Why())
	}
	if check := declarations.ValidateQualifierDiscriminators(g); !check.OK() {
		t.Fatal(check.Why())
	}
	before := g.Trace()
	p1, p2 := c.newPopulator(), c.newPopulator()
	if p1.Grammar != g || p2.Grammar != g {
		t.Fatal("populators did not receive the compiled graph")
	}
	a, err := p1.Populate(capture.Type[struct{}]())
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Populate(struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	cModel, err := c.Build[struct{}]()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || b == cModel || a == cModel {
		t.Fatal("Populate reused a model")
	}
	for _, model := range []*handoffModel{a, b, cModel} {
		if model.First.Flag.Name.Get() != "" || model.First.Flag.Name.Present() || model.First.Flag.Enabled.Present() {
			t.Fatal("fresh model has populated read state")
		}
	}
	if c.grammar != g || g.Trace() != before {
		t.Fatal("population changed the compiled graph")
	}
}

func TestFailedCompilationReturnsNoGraph(t *testing.T) {
	if c, err := Compile[badHandoffLanes](); err == nil || c != nil {
		t.Fatalf("lane: chart=%v error=%v", c, err)
	}
	if c, err := Compile[collisionModel](); err == nil || c != nil {
		t.Fatalf("namespace: chart=%v error=%v", c, err)
	}
	e, err := exploration.New()
	if err != nil {
		t.Fatal(err)
	}
	if g, err := e.Explore[badHandoffLanes](); err == nil || g != nil {
		t.Fatalf("lane: graph=%v error=%v", g, err)
	}
	if g, err := e.Explore[collisionModel](); err == nil || g != nil {
		t.Fatalf("namespace: graph=%v error=%v", g, err)
	}
	if g, err := e.Explore[struct{}](); err == nil || g != nil {
		t.Fatalf("body: graph=%v error=%v", g, err)
	}
}

func TestUncompiledNavigatorRejectsPopulation(t *testing.T) {
	var nilNavigator *Navigator[handoffModel]
	for _, navigator := range []*Navigator[handoffModel]{nilNavigator, new(Navigator[handoffModel])} {
		if model, err := navigator.Populate(struct{}{}); err == nil || model != nil {
			t.Fatalf("uncompiled Populate: model=%v error=%v", model, err)
		}
		if model, err := navigator.Build[struct{}](); err == nil || model != nil {
			t.Fatalf("uncompiled Build: model=%v error=%v", model, err)
		}
	}
}

type factoryRecipe struct {
	Flag bool `test:"flag" name:"configured"`
}

func TestRecipeIsTypeEvidenceOnly(t *testing.T) {
	c, err := Compile[handoffModel]()
	if err != nil {
		t.Fatal(err)
	}
	source := factoryRecipe{Flag: true}
	var recipe any = source
	model, err := c.Populate(recipe)
	if err != nil {
		t.Fatal(err)
	}
	if model == nil || !source.Flag {
		t.Fatal("model was not produced or Recipe was mutated")
	}
	if !model.First.Flag.Name.Present() || model.First.Flag.Name.Get() != "configured" {
		t.Fatal("Recipe tags did not populate the selected occurrence")
	}
	var typedNil *factoryRecipe
	// Typed nil still supplies exact type evidence, but pointer Recipe roots are
	// now rejected by population admission, as are nonnil pointer roots.
	if model, err := c.Populate(typedNil); err == nil || model != nil {
		t.Fatalf("pointer root admission: model=%v error=%v", model, err)
	}
	if model, err := c.Populate(nil); err == nil || model != nil {
		t.Fatalf("untyped nil: model=%v error=%v", model, err)
	}
	// The explicit entry point needs no Recipe value at all.
	if model, err := c.Build[factoryRecipe](); err != nil || model == nil {
		t.Fatalf("explicit type: model=%v error=%v", model, err)
	}
}
