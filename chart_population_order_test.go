package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/intel/rules/typerule"
	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type orderFirst struct {
	legend.Landmark `sail:"landmark" family:"first"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}
type orderSecond struct {
	legend.Landmark `sail:"landmark" family:"second"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}
type orderFirstLane struct {
	legend.Lane
	First orderFirst `sail:"spot" spot:"single,required"`
}
type orderSecondLane struct {
	legend.Lane
	Second orderSecond `sail:"spot" spot:"single,required"`
}
type orderLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	First         orderFirstLane  `sail:"lane"`
	Second        orderSecondLane `sail:"lane"`
}
type OrderEmbedded struct {
	First bool `order:"first" name:"promoted"`
}
type orderRecipe struct {
	Before  string
	Unknown int `order:"unrelated"`
	OrderEmbedded
	Between *int
	Other   int  `order:"unrelated"`
	Second  bool `order:"second" name:"direct"`
}
type orderReversed struct {
	Second bool `order:"second"`
	OrderEmbedded
}
type orderMissingSecond struct{ OrderEmbedded }

type orderUnorderedLane struct {
	legend.Lane
	First  orderFirst  `sail:"spot" spot:"single,required"`
	Second orderSecond `sail:"spot" spot:"single,required"`
}

type orderUnorderedLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Members       orderUnorderedLane `sail:"lane"`
}

func TestPopulationMembersWithinLaneAreUnordered(t *testing.T) {
	c, err := sail.Compile[orderUnorderedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	if model, err := c.Build[orderRecipe](); model == nil || err != nil {
		t.Fatalf("first then second: model=%v error=%v", model, err)
	}
	if model, err := c.Build[orderReversed](); model == nil || err != nil {
		t.Fatalf("second then first: model=%v error=%v", model, err)
	}
}

func TestPopulationOrderedPromotedFields(t *testing.T) {
	c, err := sail.Compile[orderLegend]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	m, err := c.Build[orderRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: model=%v error=%v", m, err)
	}
	if f := m.First.First.Name; !f.Present() || f.Get() != "promoted" {
		t.Fatalf("first name=(%q,%v)", f.Get(), f.Present())
	}
	if f := m.Second.Second.Name; !f.Present() || f.Get() != "direct" {
		t.Fatalf("second name=(%q,%v)", f.Get(), f.Present())
	}
	if m, err := c.Build[orderReversed](); err == nil || m != nil {
		t.Fatalf("reversed: model=%v error=%v", m, err)
	}
	if m, err := c.Build[orderMissingSecond](); err == nil || m != nil {
		t.Fatalf("missing later required lane: model=%v error=%v", m, err)
	}
}

type orderBool struct {
	legend.Landmark `sail:"landmark" family:"flag"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}

func (orderBool) Accept() typerule.RuleAPI { return typerule.Type().Is[bool]() }

type orderBoolLane struct {
	legend.Lane
	Flag orderBool `sail:"spot" spot:"single,required"`
}
type orderRegion struct {
	legend.Region `sail:"region" family:"group"`
	Name          legend.Feature    `sail:"feature" symbol:"name"`
	Flags         orderBoolLane     `sail:"lane" exploration:"eager"`
	Fallback      orderFallbackLane `sail:"lane"`
}
type orderFallback struct {
	legend.Landmark `sail:"landmark" family:"flag"`
}
type orderFallbackLane struct {
	legend.Lane
	Flag orderFallback `sail:"spot" spot:"single"`
}
type orderRegionLane struct {
	legend.Lane
	Group orderRegion `sail:"spot" spot:"single,required"`
}
type orderNestedLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Groups        orderRegionLane `sail:"lane"`
}
type orderInner struct {
	Flag bool `order:"flag" name:"inside"`
}
type orderNestedRecipe struct {
	Group orderInner `order:"group" name:"outside"`
}
type orderBadInner struct {
	Flag string `order:"flag" name:"invalid"`
}
type orderBadNestedRecipe struct {
	Group orderBadInner `order:"group" name:"bad"`
}

type orderDoubleInner struct {
	First  bool `order:"flag"`
	Second bool `order:"flag"`
}

type orderDoubleNestedRecipe struct {
	Group orderDoubleInner `order:"group"`
}

func TestPopulationFullSingleDoesNotYieldToLaterLane(t *testing.T) {
	c, err := sail.Compile[orderNestedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	// The first inner Lane is eager and its flag Spot is required/single. The
	// second matching field remains owned by that Lane despite the matching
	// optional flag Spot in the fallback Lane.
	if model, err := c.Build[orderDoubleNestedRecipe](); model != nil || err == nil || !strings.Contains(err.Error(), "cardinality violation") || !strings.Contains(err.Error(), "Second") {
		t.Fatalf("full single must fail, not yield: model=%v error=%v", model, err)
	}
}

type orderEmptyInner struct{ Other int }
type orderEmptyNestedRecipe struct {
	Group orderEmptyInner `order:"group" name:"empty"`
}

func TestPopulationNestedAcceptanceAndReuse(t *testing.T) {
	c, err := sail.Compile[orderNestedLegend]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	first, err := c.Build[orderNestedRecipe]()
	if err != nil || first == nil {
		t.Fatalf("first Build: model=%v error=%v", first, err)
	}
	if m, err := c.Build[orderBadNestedRecipe](); err == nil || m != nil || !strings.Contains(err.Error(), "acceptance") {
		t.Fatalf("selected invalid flag must not fall back: model=%v error=%v", m, err)
	}
	if m, err := c.Build[orderEmptyNestedRecipe](); err == nil || m != nil {
		t.Fatalf("missing inner flag: model=%v error=%v", m, err)
	}
	again, err := c.Populate(orderNestedRecipe{})
	if err != nil || again == nil {
		t.Fatalf("Populate after failure: model=%v error=%v", again, err)
	}
	if first == again {
		t.Fatal("reused model")
	}
	for _, m := range []*orderNestedLegend{first, again} {
		group := m.Groups.Group
		if !group.Name.Present() || group.Name.Get() != "outside" {
			t.Fatalf("region name=(%q,%v)", group.Name.Get(), group.Name.Present())
		}
		f := group.Flags.Flag.Name
		if !f.Present() || f.Get() != "inside" {
			t.Fatalf("inner name=(%q,%v)", f.Get(), f.Present())
		}
	}
}

type orderSeparator struct {
	legend.Landmark `sail:"landmark" family:"separator"`
}
type orderSeparatorLane struct {
	legend.Lane
	Separator orderSeparator `sail:"spot" spot:"single,required"`
}
type orderSharedRegion struct {
	legend.Region `sail:"region" family:"container"`
	Before        orderBoolLane      `sail:"lane" exploration:"eager"`
	Separator     orderSeparatorLane `sail:"lane"`
	After         orderBoolLane      `sail:"lane"`
}
type orderSharedRegionLane struct {
	legend.Lane
	Container orderSharedRegion `sail:"spot" spot:"single,required"`
}
type orderSharedLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Before        orderBoolLane         `sail:"lane" exploration:"eager"`
	Separator     orderSeparatorLane    `sail:"lane"`
	After         orderBoolLane         `sail:"lane"`
	Container     orderSharedRegionLane `sail:"lane"`
}
type orderSharedInner struct {
	Before    bool `order:"flag" name:"inner-before"`
	Separator bool `order:"separator"`
	After     bool `order:"flag" name:"inner-after"`
}
type orderSharedRecipe struct {
	Before    bool             `order:"flag" name:"outer-before"`
	Separator bool             `order:"separator"`
	After     bool             `order:"flag" name:"outer-after"`
	Container orderSharedInner `order:"container"`
}
type orderSharedMissingAfter struct {
	Before    bool             `order:"flag" name:"outer-before"`
	Separator bool             `order:"separator"`
	Container orderSharedInner `order:"container"`
}

func TestPopulationSharedLaneOccurrences(t *testing.T) {
	c, err := sail.Compile[orderSharedLegend]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	if m, err := c.Build[orderSharedMissingAfter](); err == nil || m != nil {
		t.Fatalf("missing required second shared use: model=%v error=%v", m, err)
	}
	m, err := c.Build[orderSharedRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: model=%v error=%v", m, err)
	}
	for _, tc := range []struct {
		name string
		f    legend.Feature
		want string
	}{
		{"outer before", m.Before.Flag.Name, "outer-before"},
		{"outer after", m.After.Flag.Name, "outer-after"},
		{"inner before", m.Container.Container.Before.Flag.Name, "inner-before"},
		{"inner after", m.Container.Container.After.Flag.Name, "inner-after"},
	} {
		if !tc.f.Present() || tc.f.Get() != tc.want {
			t.Errorf("%s=(%q,%v), want %q", tc.name, tc.f.Get(), tc.f.Present(), tc.want)
		}
	}
}
