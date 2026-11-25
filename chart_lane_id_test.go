package sail

import (
	"strings"
	"testing"

	"github.com/qlustra/sail/legend"
)

type idFlags struct {
	legend.Lane
	Flag struct {
		legend.Landmark `sail:"landmark" family:"flag"`
	} `sail:"spot" spot:"single"`
}

type idLegend struct {
	legend.Legend `sail:"legend" mark:"cli"`
	Entry         struct {
		legend.Lane
		Command struct {
			legend.Region `sail:"region" family:"command"`
			Global        idFlags `sail:"lane" lane:"global-flags" exploration:"eager"`
			Local         idFlags `sail:"lane" lane:"local-flags"`
		} `sail:"spot" spot:"single"`
	} `sail:"lane" lane:"global-flags"`
}

func TestLaneIDsBelongToUsesWithinEachBody(t *testing.T) {
	c, err := Compile[idLegend]()
	if err != nil {
		t.Fatal(err)
	}
	entry := c.grammar.Lanes[0]
	flags := entry.Declaration.Spots()[0].Region.Declaration.Lanes
	if entry.ID != "global-flags" || flags[0].ID != "global-flags" || flags[1].ID != "local-flags" {
		t.Fatalf("unexpected lane IDs: %q, %q, %q", entry.ID, flags[0].ID, flags[1].ID)
	}
	if flags[0].Declaration != flags[1].Declaration {
		t.Fatal("naming individual Lane uses should not split their shared body")
	}
}

func TestCompileRejectsDuplicateLaneIDs(t *testing.T) {
	type duplicateUses struct {
		First  idFlags `sail:"lane" lane:"flags" exploration:"eager"`
		Second idFlags `sail:"lane" lane:"flags"`
	}
	type top struct {
		legend.Legend `sail:"legend" mark:"cli"`
		duplicateUses
	}
	type nested struct {
		legend.Legend `sail:"legend" mark:"cli"`
		Entry         struct {
			legend.Lane
			Command struct {
				legend.Region `sail:"region" family:"command"`
				duplicateUses
			} `sail:"spot" spot:"single"`
		} `sail:"lane"`
	}
	if _, err := Compile[top](); err == nil || !strings.Contains(err.Error(), `fields "First" and "Second" use the same lane ID "flags"`) {
		t.Fatalf("expected duplicate lane ID diagnostic, got %v", err)
	}
	if _, err := Compile[nested](); err == nil {
		t.Fatal("duplicate lane ID in Region was accepted")
	}
}

func TestExplicitEmptyLaneIDIsNotAnUnnamedLane(t *testing.T) {
	type empty struct {
		legend.Legend `sail:"legend" mark:"cli"`
		Flags         idFlags `sail:"lane" lane:""`
	}
	if _, err := Compile[empty](); err == nil || !strings.Contains(err.Error(), "empty lane ID") {
		t.Fatalf("expected empty lane ID diagnostic, got %v", err)
	}
}
