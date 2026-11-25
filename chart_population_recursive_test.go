package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type recursiveFlagsLane struct {
	legend.Lane
	Flags []orderBool `sail:"spot" spot:"multi,required"`
}

type recursiveRegion struct {
	legend.Region `sail:"region" family:"node"`
	Name          legend.Feature        `sail:"feature" symbol:"name"`
	Flags         recursiveFlagsLane    `sail:"lane"`
	Children      recursiveChildrenLane `sail:"lane"`
}

type recursiveChildrenLane struct {
	legend.Lane
	Nodes []recursiveRegion `sail:"spot" spot:"multi"`
}

type recursiveRootLane struct {
	legend.Lane
	Nodes []recursiveRegion `sail:"spot" spot:"multi,required"`
}

type recursiveLegend struct {
	legend.Legend `sail:"legend" mark:"order"`
	Roots         recursiveRootLane `sail:"lane"`
}

type recursiveLeafRecipe struct {
	Flag bool `order:"flag" name:"leaf-flag"`
}

type recursiveChildRecipe struct {
	Flag       bool                `order:"flag" name:"child-a-flag"`
	Extra      bool                `order:"flag" name:"child-a-extra"`
	Grandchild recursiveLeafRecipe `order:"node" name:"grandchild"`
}

type recursiveFirstRootRecipe struct {
	Flag   bool                 `order:"flag" name:"root-a-flag"`
	First  recursiveChildRecipe `order:"node" name:"child-a"`
	Second recursiveLeafRecipe  `order:"node" name:"child-b"`
}

type recursiveRecipe struct {
	First  recursiveFirstRootRecipe `order:"node" name:"root-a"`
	Second recursiveLeafRecipe      `order:"node" name:"root-b"`
}

type recursiveMissingFlagsRecipe struct {
	Child recursiveLeafRecipe `order:"node" name:"unreachable-child"`
}

type recursiveBadRootRecipe struct {
	Flag bool                        `order:"flag" name:"parent-flag"`
	Good recursiveLeafRecipe         `order:"node" name:"good-sibling"`
	Bad  recursiveMissingFlagsRecipe `order:"node" name:"missing-flags"`
}

type recursiveBadRecipe struct {
	Root recursiveBadRootRecipe `order:"node" name:"bad-root"`
}

func checkRecursiveNode(t *testing.T, node recursiveRegion, name string, flags []string, children int) {
	t.Helper()
	if !node.Name.Present() || node.Name.Get() != name {
		t.Fatalf("node name=(%q,%v), want %q", node.Name.Get(), node.Name.Present(), name)
	}
	if len(node.Flags.Flags) != len(flags) || len(node.Children.Nodes) != children {
		t.Fatalf("node %s: flags=%d children=%d, want %d and %d", name, len(node.Flags.Flags), len(node.Children.Nodes), len(flags), children)
	}
	for i, want := range flags {
		flag := node.Flags.Flags[i].Name
		if !flag.Present() || flag.Get() != want {
			t.Fatalf("node %s flag %d=(%q,%v), want %q", name, i, flag.Get(), flag.Present(), want)
		}
	}
}

func checkRecursiveLegend(t *testing.T, populatedLegend *recursiveLegend) {
	t.Helper()
	if populatedLegend == nil || len(populatedLegend.Roots.Nodes) != 2 {
		t.Fatalf("expected two root nodes: %#v", populatedLegend)
	}
	roots := populatedLegend.Roots.Nodes
	checkRecursiveNode(t, roots[0], "root-a", []string{"root-a-flag"}, 2)
	checkRecursiveNode(t, roots[1], "root-b", []string{"leaf-flag"}, 0)
	children := roots[0].Children.Nodes
	checkRecursiveNode(t, children[0], "child-a", []string{"child-a-flag", "child-a-extra"}, 1)
	checkRecursiveNode(t, children[1], "child-b", []string{"leaf-flag"}, 0)
	checkRecursiveNode(t, children[0].Children.Nodes[0], "grandchild", []string{"leaf-flag"}, 0)
}

func TestRecursivePopulationSharedRegion(t *testing.T) {
	chart, err := sail.Compile[recursiveLegend]()
	if err != nil || chart == nil {
		t.Fatalf("Compile recursive graph: chart=%v error=%v", chart, err)
	}
	builtLegend, err := chart.Build[recursiveRecipe]()
	if err != nil || builtLegend == nil {
		t.Fatalf("Build: legend=%v error=%v", builtLegend, err)
	}
	populatedLegend, err := chart.Populate(recursiveRecipe{})
	if err != nil || populatedLegend == nil {
		t.Fatalf("Populate: legend=%v error=%v", populatedLegend, err)
	}
	checkRecursiveLegend(t, builtLegend)
	checkRecursiveLegend(t, populatedLegend)

	if failedLegend, err := chart.Build[recursiveBadRecipe](); err == nil || failedLegend != nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("missing deeper flags: legend=%v error=%v", failedLegend, err)
	}
	againLegend, err := chart.Build[recursiveRecipe]()
	if err != nil || againLegend == nil {
		t.Fatalf("Build after failure: legend=%v error=%v", againLegend, err)
	}
	checkRecursiveLegend(t, againLegend)

	children := againLegend.Roots.Nodes[0].Children.Nodes
	children[0].Flags.Flags[0] = children[0].Flags.Flags[1]
	checkRecursiveNode(t, children[1], "child-b", []string{"leaf-flag"}, 0)
	checkRecursiveNode(t, children[0].Children.Nodes[0], "grandchild", []string{"leaf-flag"}, 0)
	checkRecursiveNode(t, againLegend.Roots.Nodes[1], "root-b", []string{"leaf-flag"}, 0)
	children[0] = children[1]
	checkRecursiveNode(t, againLegend.Roots.Nodes[1], "root-b", []string{"leaf-flag"}, 0)
	checkRecursiveLegend(t, builtLegend)
	checkRecursiveLegend(t, populatedLegend)
}
