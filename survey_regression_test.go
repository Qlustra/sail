package sail_test

import (
	"reflect"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type surveyLegend struct {
	legend.Legend `sail:"legend" mark:"test" legend:"survey"`
	Items         surveyLane `sail:"lane"`
}

type surveyLane struct {
	legend.Lane
	Leaves []surveyLeaf  `sail:"spot" spot:"multi"`
	Groups []surveyGroup `sail:"spot" spot:"multi"`
}

type surveyGroup struct {
	legend.Region `sail:"region" family:"group"`
	Name          legend.Feature `sail:"feature" symbol:"name"`
	Items         surveyLane     `sail:"lane"`
}

type surveyLeaf struct {
	legend.Landmark `sail:"landmark" family:"leaf"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}

type surveyChart struct {
	First struct {
		Child string `test:"leaf" name:"one"`
	} `test:"group" name:"first"`
	Middle string `test:"leaf" name:"middle"`
	Last   struct {
		Child string `test:"leaf" name:"two"`
	} `test:"group" name:"last"`
}

func TestSurveyOrderAndReducerIsolation(t *testing.T) {
	nav, err := sail.Compile[surveyLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := nav.Build[surveyChart]()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		order sail.TraversalOrder
		want  []string
	}{
		{sail.ChartOrder, []string{"first:seed", "first+:seed/first", "one:seed/first+", "middle:seed", "last:seed", "last+:seed/last", "two:seed/last+"}},
		{sail.LegendOrder, []string{"middle:seed", "first:seed", "first+:seed/first", "one:seed/first+", "last:seed", "last+:seed/last", "two:seed/last+"}},
	} {
		t.Run(string(tc.order), func(t *testing.T) {
			var got []string
			group := sail.Select[*surveyGroup]().Encounter(func(_ sail.Encounter, prefix string, token *surveyGroup) string {
				got = append(got, token.Name.Get()+":"+prefix)
				return prefix + "/" + token.Name.Get()
			})
			second := sail.Select[*surveyGroup]().Encounter(func(_ sail.Encounter, prefix string, token *surveyGroup) string {
				got = append(got, token.Name.Get()+"+:"+prefix)
				return prefix + "+"
			})
			leaf := sail.Select[*surveyLeaf]().Encounter(func(_ sail.Encounter, prefix string, token *surveyLeaf) string {
				got = append(got, token.Name.Get()+":"+prefix)
				return "replacement"
			})
			if err := nav.Survey(model, tc.order, "seed", group, second, leaf); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("encounters = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSurveyPrefixAssignmentSharesReferences(t *testing.T) {
	nav, err := sail.Compile[surveyLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := nav.Build[surveyChart]()
	if err != nil {
		t.Fatal(err)
	}
	seed := map[string]string{}
	var observed string
	rule := sail.Select[*surveyLeaf]().Encounter(func(_ sail.Encounter, prefix map[string]string, token *surveyLeaf) map[string]string {
		switch token.Name.Get() {
		case "one":
			prefix["shared"] = "visible"
			return map[string]string{"shared": "replacement"}
		case "middle":
			observed = prefix["shared"]
		}
		return prefix
	})
	if err := nav.Survey(model, sail.ChartOrder, seed, rule); err != nil {
		t.Fatal(err)
	}
	if observed != "visible" || seed["shared"] != "visible" {
		t.Fatalf("sibling observed %q, seed = %v; want shared mutation without replacement", observed, seed)
	}
}
