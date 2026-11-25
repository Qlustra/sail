package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type graphAB struct {
	legend.Landmark `sail:"landmark" family:"graphflag" discriminate:"alpha,beta"`
}
type graphBA struct {
	legend.Landmark `sail:"landmark" family:"graphflag" discriminate:"beta,alpha"`
}
type graphOther struct {
	legend.Landmark `sail:"landmark" family:"graphflag" discriminate:"other"`
}
type graphEagerSelectors struct {
	legend.Lane
	Flag graphAB `sail:"spot" spot:"single"`
}
type graphRequiredSame struct {
	legend.Lane
	Flag graphBA `sail:"spot" spot:"single,required"`
}
type graphRequiredOther struct {
	legend.Lane
	Flag graphOther `sail:"spot" spot:"single,required"`
}
type graphSameSelectors struct {
	legend.Legend `sail:"legend" mark:"graphsame"`
	Eager         graphEagerSelectors `sail:"lane" exploration:"eager"`
	Later         graphRequiredSame   `sail:"lane"`
}
type graphDistinctSelectors struct {
	legend.Legend `sail:"legend" mark:"graphdistinct"`
	Eager         graphEagerSelectors `sail:"lane" exploration:"eager"`
	Later         graphRequiredOther  `sail:"lane"`
}

func TestLaneConsumptionGraphNormalizedSelectors(t *testing.T) {
	if c, err := sail.Compile[graphSameSelectors](); c != nil || err == nil {
		t.Fatalf("equal selector sets must starve required spot: chart=%v error=%v", c, err)
	} else if !strings.Contains(strings.ToLower(err.Error()), "unreachable") {
		t.Fatalf("expected unreachable requirement, got %v", err)
	}
	c, err := sail.Compile[graphDistinctSelectors]()
	if err != nil || c == nil {
		t.Fatalf("distinct selector is an entry boundary: chart=%v error=%v", c, err)
	}
	if len(c.Warnings()) != 0 {
		t.Fatalf("distinct required selector is not shadowed: %#v", c.Warnings())
	}
}

// The recursive link and both root uses point to the same compiled Region body.
type graphSharedRegion struct {
	legend.Region `sail:"region" family:"sharedgraph"`
	Eager         consumptionOptionalFlag    `sail:"lane" exploration:"eager"`
	Later         consumptionBoundaryAndFlag `sail:"lane"`
	Children      graphChildren              `sail:"lane"`
}
type graphChildren struct {
	legend.Lane
	Self []graphSharedRegion `sail:"spot" spot:"multi"`
}
type graphSharedUses struct {
	legend.Lane
	FirstShared graphSharedRegion `sail:"spot" spot:"single"`
}
type graphSharedUsesAgain struct {
	legend.Lane
	SecondShared graphSharedRegion `sail:"spot" spot:"single"`
}
type graphSharedRoot struct {
	legend.Legend `sail:"legend" mark:"sharedroot"`
	FirstHost     graphSharedUses      `sail:"lane" exploration:"eager"`
	SecondHost    graphSharedUsesAgain `sail:"lane"`
}

type graphRegionA struct {
	legend.Region `sail:"region" family:"grapha"`
	Eager         consumptionOptionalFlag    `sail:"lane" exploration:"eager"`
	Later         consumptionBoundaryAndFlag `sail:"lane"`
}
type graphRegionB struct {
	legend.Region `sail:"region" family:"graphb"`
	Eager         consumptionOptionalFlag    `sail:"lane" exploration:"eager"`
	Later         consumptionBoundaryAndFlag `sail:"lane"`
}
type graphTwoRegions struct {
	legend.Lane
	AlphaSite graphRegionA `sail:"spot" spot:"single"`
	BetaSite  graphRegionB `sail:"spot" spot:"single"`
}
type graphTwoRoot struct {
	legend.Legend `sail:"legend" mark:"tworoot"`
	HostSites     graphTwoRegions `sail:"lane"`
}
type graphDeadRegion struct {
	legend.Region `sail:"region" family:"deadgraph"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Later         orderBoolLane           `sail:"lane"`
}
type graphDeadUses struct {
	legend.Lane
	DeadSite graphDeadRegion `sail:"spot" spot:"single"`
}
type graphDeadRoot struct {
	legend.Legend `sail:"legend" mark:"deadroot"`
	DeadHost      graphDeadUses `sail:"lane"`
}

func TestLaneConsumptionGraphSharedRecursiveBody(t *testing.T) {
	c, err := sail.Compile[graphSharedRoot]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	w := c.Warnings()
	if len(w) != 1 || w[0].Code != "eager_lane_entry_boundary" {
		t.Fatalf("shared body must yield one warning: %#v", w)
	}
	graphMessageContains(t, w[0].Message, "sharedroot", "FirstHost", "FirstShared", "Eager", "Later")
}

func TestLaneConsumptionGraphDistinctBodyProvenance(t *testing.T) {
	c, err := sail.Compile[graphTwoRoot]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	w := c.Warnings()
	if len(w) != 2 || w[0].Code != "eager_lane_entry_boundary" || w[1].Code != "eager_lane_entry_boundary" || w[0].Message == w[1].Message {
		t.Fatalf("distinct bodies must yield distinct warnings: %#v", w)
	}
	for i, site := range []string{"AlphaSite", "BetaSite"} {
		graphMessageContains(t, w[i].Message, "tworoot", "HostSites", site, "Eager", "Later", "Flag")
	}
}

func TestLaneConsumptionGraphDeadBodyProvenance(t *testing.T) {
	c, err := sail.Compile[graphDeadRoot]()
	if c != nil || err == nil {
		t.Fatalf("expected dead nested body to fail: chart=%v error=%v", c, err)
	}
	graphMessageContains(t, err.Error(), "deadroot", "DeadHost", "DeadSite", "Eager", "Later", "required", "unreachable")
}

type graphBridgeBoundary struct {
	legend.Lane
	BridgeExit orderSeparator `sail:"spot" spot:"single"`
}
type graphLaterBoundary struct {
	legend.Lane
	LaterExit orderSeparator `sail:"spot" spot:"single"`
	Flag      orderBool      `sail:"spot" spot:"single,required"`
}
type graphBridgeRoot struct {
	legend.Legend `sail:"legend" mark:"bridgeroot"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Bridge        graphBridgeBoundary     `sail:"lane" exploration:"eager"`
	Later         graphLaterBoundary      `sail:"lane"`
}

func TestLaneConsumptionGraphBoundarySelectorGuidance(t *testing.T) {
	c, err := sail.Compile[graphBridgeRoot]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	w := c.Warnings()
	if len(w) != 1 || w[0].Code != "eager_lane_entry_boundary" {
		t.Fatalf("expected one boundary warning: %#v", w)
	}
	graphMessageContains(t, w[0].Message, "separator", "Eager", "Later")
	for _, spot := range []string{"BridgeExit", "LaterExit"} {
		if strings.Contains(strings.ToLower(w[0].Message), strings.ToLower(spot)) {
			t.Errorf("selector guidance must not attribute boundary to spot %s: %s", spot, w[0].Message)
		}
	}
	if strings.Count(strings.ToLower(w[0].Message), "separator") != 1 {
		t.Errorf("repeated boundary selector must be deduplicated: %s", w[0].Message)
	}
}

func graphMessageContains(t *testing.T, message string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(strings.ToLower(message), strings.ToLower(part)) {
			t.Errorf("message %q does not contain %q", message, part)
		}
	}
}
