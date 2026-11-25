package traversal

import (
	"reflect"
	"strings"
	"testing"

	"github.com/qlustra/sail/internal/chart/exploration"
	"github.com/qlustra/sail/legend"
)

type prepareLeaf struct {
	legend.Landmark `sail:"landmark" family:"leaf"`
}
type prepareRegion struct {
	legend.Region `sail:"region" family:"region"`
	Children      prepareLane `sail:"lane" lane:"children"`
}
type prepareLane struct {
	legend.Lane
	Region []prepareRegion `sail:"spot" spot:"multi"`
	Leaves prepareLeaf     `sail:"spot" spot:"single"`
}
type prepareLegend struct {
	legend.Legend `sail:"legend" mark:"test"`
	First         prepareLane `sail:"lane" lane:"first" exploration:"eager"`
	Second        prepareLane `sail:"lane" lane:"second"`
}

func TestValidateWaypoints(t *testing.T) {
	explorer, err := exploration.New()
	if err != nil {
		t.Fatal(err)
	}
	grammar, err := explorer.Explore[prepareLegend]()
	if err != nil {
		t.Fatal(err)
	}
	region := reflect.TypeFor[*prepareRegion]()
	leaf := reflect.TypeFor[*prepareLeaf]()
	anyType := reflect.TypeFor[any]()
	type distinctPointer *prepareLeaf
	tests := []struct {
		name    string
		steps   []WaypointShape
		failure string
	}{
		{"empty", nil, ""},
		{"recursive and multi", []WaypointShape{{region, "second"}, {region, "children"}, {leaf, "children"}}, ""},
		{"interface branch", []WaypointShape{{anyType, "first"}, {leaf, "children"}}, ""},
		{"wrong lane", []WaypointShape{{region, "missing"}}, "waypoint 1"},
		{"value not pointer", []WaypointShape{{reflect.TypeFor[prepareRegion](), ""}}, "waypoint 1"},
		{"assignable but not identical", []WaypointShape{{reflect.TypeFor[distinctPointer](), ""}}, "waypoint 1"},
		{"terminal", []WaypointShape{{leaf, "first"}, {region, "children"}}, "terminal Landmark"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWaypoints(grammar, tt.steps)
			if tt.failure == "" && err != nil {
				t.Fatal(err)
			}
			if tt.failure != "" && (err == nil || !strings.Contains(err.Error(), tt.failure)) {
				t.Fatalf("got %v, want %q", err, tt.failure)
			}
		})
	}
}
