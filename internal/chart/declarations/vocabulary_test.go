package declarations

import (
	"reflect"
	"sync"
	"testing"

	"github.com/qlustra/intel/define"
)

func vocabularyFixture(t *testing.T) (*Legend, *Lane, *Spot, *Spot) {
	t.Helper()
	plain := &Spot{For: SpotForLandmark, Landmark: &define.Carrier[Landmark]{Declaration: &Landmark{Family: "flag", Qualifiers: []*QualifierCarrier{{Symbol: "optional"}}}}}
	enumBody := &Region{Family: "flag", Discriminators: []string{"enum", "other"}, Qualifiers: []*QualifierCarrier{{Symbol: "required"}}}
	enum := &Spot{For: SpotForRegion, Region: &define.Carrier[Region]{Declaration: enumBody}}
	other := &Spot{For: SpotForLandmark, Landmark: &define.Carrier[Landmark]{Declaration: &Landmark{Family: "other", Discriminators: []string{"optional"}, Qualifiers: []*QualifierCarrier{{Symbol: "enum"}}}}}
	same := &Spot{For: SpotForLandmark, Landmark: &define.Carrier[Landmark]{Declaration: &Landmark{Family: "same", Discriminators: []string{"same"}}}}
	lane := &Lane{}
	for _, spot := range []*Spot{plain, enum, other, same} {
		if check := lane.AddSpot(spot); !check.OK() {
			t.Fatal(check.Why())
		}
	}
	use := &LaneCarrier{Carrier: &define.Carrier[Lane]{Declaration: lane}}
	enumBody.Lanes = []*LaneCarrier{use} // shared and recursive
	return &Legend{Lanes: []*LaneCarrier{use, use}}, lane, plain, enum
}

func TestVocabularyClassificationAndSelection(t *testing.T) {
	graph, lane, plain, enum := vocabularyFixture(t)
	v, check := BuildVocabulary(graph)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if graph.Vocabulary != nil {
		t.Fatal("building published vocabulary")
	}
	if check := ValidateQualifierDiscriminators(graph); !check.OK() || graph.Vocabulary != nil {
		t.Fatal("validator must be pure")
	}
	mark, check := v.Classify("flag", nil)
	if !check.OK() || lane.Select(mark) != plain {
		t.Fatal("plain selector")
	}
	input := []string{"required", "other", "enum"}
	mark, check = v.Classify("flag", input)
	if !check.OK() || lane.Select(mark) != enum || !mark.HasQualifier("required") || mark.HasQualifier("optional") {
		t.Fatal("enum selector or qualifier")
	}
	if !reflect.DeepEqual(input, []string{"required", "other", "enum"}) {
		t.Fatal("input mutated")
	}
	input[0] = "optional"
	if !mark.HasQualifier("required") || mark.HasQualifier("optional") {
		t.Fatal("packet retained caller slice")
	}
	if check := mark.ValidateFor(enum); !check.OK() {
		t.Fatal(check.Why())
	}
	mark, check = v.Classify("flag", []string{"optional", "enum", "other"})
	if !check.OK() || lane.Select(mark) != enum {
		t.Fatal("qualifier changed selection")
	}
	if check := mark.ValidateFor(enum); check.Code != "invalid_recipe_token" {
		t.Fatalf("unsupported qualifier: %v", check)
	}
	mark, check = v.Classify("flag", []string{"enum"})
	if !check.OK() || lane.Select(mark) != nil {
		t.Fatal("selector must be exact")
	}
	mark, check = v.Classify("other", []string{"optional", "enum"})
	if !check.OK() || !mark.HasQualifier("enum") || mark.HasQualifier("optional") {
		t.Fatal("family-specific roles")
	}
	mark, check = v.Classify("same", []string{"same"})
	if !check.OK() || lane.Select(mark) == nil {
		t.Fatal("family and modifier occupy separate namespaces")
	}
	for _, tc := range []struct {
		family string
		tokens []string
	}{
		{"flag", []string{""}}, {"flag", []string{"enum", "enum"}}, {"flag", []string{"optional", "optional"}},
		{"flag", []string{"unknown"}}, {"flag", []string{" enum"}}, {"flag", []string{"Enum"}}, {"missing", nil},
	} {
		if _, check := v.Classify(tc.family, tc.tokens); check.Code != "invalid_recipe_token" {
			t.Fatalf("%v: %v", tc, check)
		}
	}
}

func TestVocabularyConcurrentReads(t *testing.T) {
	graph, lane, _, enum := vocabularyFixture(t)
	v, check := BuildVocabulary(graph)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mark, check := v.Classify("flag", []string{"other", "required", "enum"})
				if !check.OK() || lane.Select(mark) != enum || !mark.HasQualifier("required") || !mark.ValidateFor(enum).OK() {
					t.Error("concurrent classification failed")
					return
				}
			}
		}()
	}
	wg.Wait()
}
