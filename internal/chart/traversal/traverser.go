package traversal

import (
	"fmt"
	"slices"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/values"
)

// One present Region or Landmark. The selector sees its pointer; the walk
// uses its Intel capture to enter the body if it is a Region.
type spotToken struct {
	markerValue   *capture.ValueInterrogator
	markerPointer any
	lane          *declarations.LaneCarrier
	spot          *declarations.Spot
	chartIndex    []int
}

// Survey visits the tokens of a populated Legend in depth-first order. The
// Legend is depth zero. Each immediate token of a Region inherits the prefix
// returned for that Region, independently of its siblings.
// Pointers passed to visit follow ordinary Go storage and slice lifetime rules.
func Survey[L, P any](grammar *declarations.Legend, legend *L, chartOrder bool, scopeSeed P, visit func(int, *declarations.LaneCarrier, any, P) P) error {
	if grammar == nil || visit == nil {
		return fmt.Errorf("survey requires a compiled grammar and a visitor")
	}
	if legend == nil {
		return fmt.Errorf("survey requires a non-nil pointer to the populated Legend")
	}
	potentialLegendMarker, check := capture.Value(legend)
	if !check.OK() {
		return fmt.Errorf("survey Legend: %s", check.Why())
	}
	legendMarker, check := potentialLegendMarker.AsStruct()
	if !check.OK() {
		return fmt.Errorf("survey Legend: %s", check.Why())
	}
	return walkScope[P](legendMarker, grammar.Lanes, chartOrder, 1, scopeSeed, visit)
}

func walkScope[P any](marker *capture.StructValueInterrogator, lanes []*declarations.LaneCarrier, chartOrder bool, depth int, inheritedScope P, visit func(int, *declarations.LaneCarrier, any, P) P) error {
	spotTokens, err := getBodySpotTokens(marker, lanes, chartOrder)
	if err != nil {
		return err
	}
	for _, candidateSpotToken := range spotTokens {
		prefix := visit(depth, candidateSpotToken.lane, candidateSpotToken.markerPointer, inheritedScope)
		if candidateSpotToken.spot.For == declarations.SpotForRegion {
			region, check := candidateSpotToken.markerValue.AsStruct()
			if !check.OK() {
				return fmt.Errorf("Region destination: %s", check.Why())
			}
			if err := walkScope[P](region, candidateSpotToken.spot.Region.Declaration.Lanes, chartOrder, depth+1, prefix, visit); err != nil {
				return err
			}
		}
	}
	return nil
}

func getBodySpotTokens(hostMarker *capture.StructValueInterrogator, lanes []*declarations.LaneCarrier, chartOrder bool) ([]spotToken, error) {
	var tokens []spotToken
	for _, lane := range lanes {
		laneValue, check := hostMarker.Field(lane.Property.FieldDescriptor)
		if !check.OK() {
			return nil, fmt.Errorf("Lane destination %s: %s", lane.Property.FieldName(), check.Why())
		}
		laneBody, check := laneValue.AsStruct()
		if !check.OK() {
			return nil, fmt.Errorf("Lane destination %s: %s", lane.Property.FieldName(), check.Why())
		}
		for _, spot := range lane.Declaration.Spots() {
			var spotFieldDescriptor *inspect.FieldDescriptor
			if spot.For == declarations.SpotForRegion {
				spotFieldDescriptor = spot.Region.Property.FieldDescriptor
			} else {
				spotFieldDescriptor = spot.Landmark.Property.FieldDescriptor
			}
			markerValue, check := laneBody.Field(spotFieldDescriptor)
			if !check.OK() {
				return nil, fmt.Errorf("Spot destination %s: %s", spotFieldDescriptor.FieldName(), check.Why())
			}
			markerValues := []*capture.ValueInterrogator{markerValue}
			if spot.Kind == declarations.MultiSpot {
				slice, check := markerValue.AsSlice()
				if !check.OK() {
					return nil, fmt.Errorf("Repeated Spot %s: %s", spotFieldDescriptor.FieldName(), check.Why())
				}
				markerValues, check = slice.Elements()
				if !check.OK() {
					return nil, fmt.Errorf("Repeated Spot %s: %s", spotFieldDescriptor.FieldName(), check.Why())
				}
			}
			for _, markerValue := range markerValues {
				markerPointer, check := markerValue.Address()
				if !check.OK() {
					return nil, fmt.Errorf("Spot %s address: %s", spotFieldDescriptor.FieldName(), check.Why())
				}
				location, ok := values.StoredLocation(markerPointer)
				if !ok {
					return nil, fmt.Errorf("Spot %s does not expose occurrence state", spotFieldDescriptor.FieldName())
				}
				if location == nil {
					continue
				}
				t := spotToken{markerValue: markerValue, markerPointer: markerPointer, lane: lane, spot: spot}
				if chartOrder {
					t.chartIndex = location.Index()
				}
				tokens = append(tokens, t)
			}
		}
	}
	if chartOrder {
		slices.SortStableFunc(tokens, func(a, b spotToken) int { return slices.Compare(a.chartIndex, b.chartIndex) })
	}
	return tokens, nil
}
