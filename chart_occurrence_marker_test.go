package sail_test

import (
	"reflect"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

// A defined marker retains its embedded behavior, but does not contain a field
// whose exact type is legend.Region or legend.Landmark.
type occurrenceRegionMarker legend.Region
type occurrenceLandmarkMarker legend.Landmark

type occurrenceDerivedRegion struct {
	occurrenceRegionMarker `sail:"region" family:"region"`
	Name                   legend.Feature `sail:"feature" symbol:"name"`
}

type occurrenceDerivedLandmark struct {
	occurrenceLandmarkMarker `sail:"landmark" family:"leaf"`
	Name                     legend.Feature `sail:"feature" symbol:"name"`
}

type occurrenceDerivedLane struct {
	legend.Lane
	Region occurrenceDerivedRegion   `sail:"spot" spot:"single"`
	Leaf   occurrenceDerivedLandmark `sail:"spot" spot:"single"`
}

type occurrenceDerivedLegend struct {
	legend.Legend `sail:"legend" mark:"occ"`
	Items         occurrenceDerivedLane `sail:"lane"`
}

type occurrenceDerivedRecipe struct {
	Ignored int
	Region  struct{} `occ:"region" name:"region"`
	Leaf    bool     `occ:"leaf" name:"leaf"`
}

func TestOccurrenceDefinedMarkers(t *testing.T) {
	chart, err := sail.Compile[occurrenceDerivedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[occurrenceDerivedRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	root := reflect.TypeFor[occurrenceDerivedRecipe]()
	occurrenceCheck(t, populatedLegend.Items.Region, root, []int{1}, "Region")
	occurrenceCheck(t, populatedLegend.Items.Leaf, root, []int{2}, "Leaf")
	if populatedLegend.Items.Region.Name.Get() != "region" || populatedLegend.Items.Leaf.Name.Get() != "leaf" {
		t.Fatal("defined markers hid authored sibling Features")
	}
}
