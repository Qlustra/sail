package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/sail"
)

// Recipe occurrences are properties only. Embedding tags cannot select a node,
// but visible promoted properties still participate at the embedding's position.
type PopulationSelectedInner struct {
	Flag bool `order:"flag" name:"inside"`
}

type populationTaggedEmbedding struct {
	PopulationSelectedInner `order:"group" name:"outside"`
}

type populationTaggedComposition struct {
	OrderEmbedded `order:"second" name:"not-an-occurrence"`
	Second        bool `order:"second" name:"actual-second"`
}

type populationBlankProperty struct {
	_     orderInner `order:"group" name:"not-an-occurrence"`
	Group orderInner `order:"group" name:"actual-group"`
}

type populationOtherScope struct {
	Before bool       `order:"flag" name:"also-root"`
	Group  orderInner `order:"group" name:"outside"`
	Extra  bool       `order:"flag" name:"root"`
}

func TestPopulationEmbeddingCannotSupplyOccurrence(t *testing.T) {
	c, err := sail.Compile[orderNestedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if model, err := c.Build[populationTaggedEmbedding](); model != nil || err == nil || !strings.Contains(err.Error(), "required Spot") {
		t.Fatalf("embedding cannot supply the required group: model=%v error=%v", model, err)
	}
}

func TestPopulationTaggedEmbeddingStillPromotesProperties(t *testing.T) {
	c, err := sail.Compile[orderLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := c.Build[populationTaggedComposition]()
	if model == nil || err != nil {
		t.Fatalf("tagged composition: model=%v error=%v", model, err)
	}
	if !model.First.First.Name.Present() || model.First.First.Name.Get() != "promoted" || !model.Second.Second.Name.Present() || model.Second.Second.Name.Get() != "actual-second" {
		t.Fatal("embedding tag disrupted promoted-property order or values")
	}
}

func TestPopulationBlankFieldCannotSupplyOccurrence(t *testing.T) {
	c, err := sail.Compile[orderNestedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := c.Build[populationBlankProperty]()
	if model == nil || err != nil {
		t.Fatalf("blank field: model=%v error=%v", model, err)
	}
	if model.Groups.Group.Name.Get() != "actual-group" {
		t.Fatal("blank field supplied an occurrence")
	}
}

func TestPopulationDefersDanglingMemberDiagnosis(t *testing.T) {
	c, err := sail.Compile[orderNestedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	// The root flag fields are not members of this scope. A later catalog and
	// remainder pass can diagnose them without making it a population prerequisite.
	model, err := c.Build[populationOtherScope]()
	if model == nil || err != nil {
		t.Fatalf("other-scope fields: model=%v error=%v", model, err)
	}
	if model.Groups.Group.Name.Get() != "outside" || model.Groups.Group.Flags.Flag.Name.Get() != "inside" {
		t.Fatal("ignored root fields disrupted ordinary nested population")
	}
}
