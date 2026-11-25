package populating

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/options"
	"github.com/qlustra/sail/internal/chart/carriers"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/values"
)

func spotMetadata(spot *declarations.Spot) (string, *inspect.FieldDescriptor) {
	if spot.For == declarations.SpotForRegion {
		return spot.Region.Declaration.Family, spot.Region.Property.FieldDescriptor
	}
	return spot.Landmark.Declaration.Family, spot.Landmark.Property.FieldDescriptor
}

// scope consumes Go-visible properties in embedding-slot order. Embeddings supply
// promoted properties, not occurrences of their own, even when tagged. Intel
// preserves the original paths; Sail does not manually unfold or hoist fields.
// TODO: Later dangling-member diagnostics may inspect tagged embeddings too;
// warning about inadmissible material is separate from consuming occurrences.
func (p *Populator[L]) scope(chartStruct *capture.StructInterrogator, chartScopeIndex []int, legendScope *capture.StructValueInterrogator, laneCarriers []*declarations.LaneCarrier) error {
	spotCounts := make(map[*declarations.Spot]int)
	laneIndex := 0
	finishLane := func(laneCarrier *declarations.LaneCarrier) error {
		for _, spot := range laneCarrier.Declaration.Spots() {
			if spot.Required && spotCounts[spot] == 0 {
				_, legendSpotProperty := spotMetadata(spot)
				return fmt.Errorf("Chart %s: required Spot %s.%s has no occurrence", chartStruct.TypeName(), laneCarrier.Property.FieldName(), legendSpotProperty.FieldName())
			}
		}
		return nil
	}
	for chartPropertyName, chartProperty := range chartStruct.Properties.AllAddressable(options.DepthFirst) {
		raw, present := chartProperty.Tags().Lookup(p.Grammar.Mark)
		if !present {
			continue
		}
		parts := strings.Split(raw, ",")
		// The family is the first token; modifier validation applies only in scope.
		family := parts[0]
		inScope := false
		for _, laneCarrier := range laneCarriers {
			for _, spot := range laneCarrier.Declaration.Spots() {
				spotFamily, _ := spotMetadata(spot)
				inScope = inScope || spotFamily == family
			}
		}
		if !inScope {
			// Open grammar: fields outside this scope's vocabulary do not move
			// the Lane cursor. TODO: A later dangling-member diagnostic can use a
			// Chart-wide catalog and evaluate unconsumed fields after population.
			// Catalog diagnostics must distinguish grammar recognition from
			// admitted field categories, not just recognize a global family
			// spelling; do not add that prerequisite now.
			continue
		}
		chartMark, check := p.Grammar.Vocabulary.Classify(family, parts[1:])
		if !check.OK() {
			return fmt.Errorf("Chart field %s: %s", chartPropertyName, check.Why())
		}
		var selectedSpot *declarations.Spot
		for laneIndex < len(laneCarriers) {
			selectedSpot = laneCarriers[laneIndex].Declaration.Select(chartMark)
			if selectedSpot != nil {
				break
			}
			if err := finishLane(laneCarriers[laneIndex]); err != nil {
				return err
			}
			laneIndex++                                   // Retry this same candidate in the next Lane.
			spotCounts = make(map[*declarations.Spot]int) // A shared Lane body is a fresh use here.
		}
		if selectedSpot == nil {
			return fmt.Errorf("Chart field %s: selector %s does not match the remaining grammar", chartPropertyName, chartMark)
		}
		if selectedSpot.Kind == declarations.SingleSpot && spotCounts[selectedSpot] != 0 {
			// Ownership is determined before policy validation. A full single
			// Spot is a violation, never a selector miss that advances the Lane.
			_, legendSpotProperty := spotMetadata(selectedSpot)
			return fmt.Errorf("Chart field %s: cardinality violation at single Spot %s.%s; an occurrence is already present", chartPropertyName, laneCarriers[laneIndex].Property.FieldName(), legendSpotProperty.FieldName())
		}
		if check := chartMark.ValidateFor(selectedSpot); !check.OK() {
			return fmt.Errorf("Chart field %s: %s", chartPropertyName, check.Why())
		}
		if check := chartStruct.FieldPathIsValue(chartProperty.FieldDescriptor); !check.OK() {
			return fmt.Errorf("Chart field %s must have a value-only path: %s", chartPropertyName, check.Why())
		}
		if chartProperty.FieldTypeInspector().IsPointerKind() {
			return fmt.Errorf("Chart field %s must be a value, not a pointer", chartPropertyName)
		}
		legendLaneValue, check := legendScope.Field(laneCarriers[laneIndex].Property.FieldDescriptor)
		if !check.OK() {
			return fmt.Errorf("Lane destination %s: %s", laneCarriers[laneIndex].Property.FieldName(), check.Why())
		}
		legendLane, check := legendLaneValue.AsStruct()
		if !check.OK() {
			return fmt.Errorf("Lane destination: %s", check.Why())
		}
		_, legendSpotProperty := spotMetadata(selectedSpot)
		legendSpotValue, check := legendLane.Field(legendSpotProperty)
		if !check.OK() {
			return fmt.Errorf("Spot destination %s: %s", legendSpotProperty.FieldName(), check.Why())
		}
		legendOccurrenceValue := legendSpotValue
		if selectedSpot.Kind == declarations.MultiSpot {
			// Each selected Chart property contributes one occurrence. This
			// does not expand collections on the Chart side.
			legendOccurrences, check := legendSpotValue.AsSlice()
			if !check.OK() {
				return fmt.Errorf("Repeated Spot destination %s: %s", legendSpotProperty.FieldName(), check.Why())
			}
			legendOccurrenceValue, check = legendOccurrences.AppendZero()
			if !check.OK() {
				return fmt.Errorf("Repeated Spot occurrence %s: %s", legendSpotProperty.FieldName(), check.Why())
			}
		}
		legendOccurrence, check := legendOccurrenceValue.AsStruct()
		if !check.OK() {
			return fmt.Errorf("Spot destination %s: %s", legendSpotProperty.FieldName(), check.Why())
		}
		// Finish this occurrence before another append can move the backing array.
		// Failed population returns no Legend; there is no partial-model contract.
		chartOccurrenceIndex := slices.Concat(chartScopeIndex, chartProperty.Index())
		if err := p.occurrence(selectedSpot, chartProperty, chartMark, chartOccurrenceIndex, legendOccurrence); err != nil {
			return fmt.Errorf("Chart field %s: %w", chartPropertyName, err)
		}
		legendOccurrenceAddress, check := legendOccurrenceValue.Address()
		if !check.OK() {
			return fmt.Errorf("Chart field %s occurrence address: %s", chartPropertyName, check.Why())
		}
		if check := values.RecordOccurrence(legendOccurrenceAddress, p.chartType, chartOccurrenceIndex); !check.OK() {
			return fmt.Errorf("Chart field %s occurrence location: %s", chartPropertyName, check.Why())
		}
		spotCounts[selectedSpot]++
	}
	for ; laneIndex < len(laneCarriers); laneIndex++ {
		if err := finishLane(laneCarriers[laneIndex]); err != nil {
			return err
		}
		spotCounts = make(map[*declarations.Spot]int)
	}
	return nil
}

func (p *Populator[L]) occurrence(spot *declarations.Spot, chartProperty *capture.PropertyInterrogator, chartMark *declarations.ChartMark, chartOccurrenceIndex []int, legendOccurrence *capture.StructValueInterrogator) error {
	chartCarrierType := capture.TypeInspector(chartProperty.FieldTypeInspector())
	var featureCarriers []*carriers.Feature
	var barrelCarriers []*carriers.Barrel
	var qualifierCarriers []*declarations.QualifierCarrier
	var chartStruct *capture.StructInterrogator
	if spot.For == declarations.SpotForRegion {
		// Admit the Region's struct value before invoking declaration providers.
		// Descent does not unwrap pointers or expand Chart collections.
		var check logic.Check
		chartStruct, check = chartCarrierType.AsStruct()
		if !check.OK() {
			return fmt.Errorf("Region Chart must be a struct value: %s", check.Why())
		}
		featureCarriers, barrelCarriers = spot.Region.Declaration.Features, spot.Region.Declaration.Barrels
		qualifierCarriers = spot.Region.Declaration.Qualifiers
	} else {
		landmarkDeclaration := spot.Landmark.Declaration
		if check := AcceptLandmark(landmarkDeclaration, chartCarrierType); !check.OK() {
			return fmt.Errorf("Landmark acceptance: %s", check.Why())
		}
		featureCarriers, barrelCarriers = landmarkDeclaration.Features, landmarkDeclaration.Barrels
		qualifierCarriers = landmarkDeclaration.Qualifiers
	}
	for _, featureCarrier := range featureCarriers {
		chartTagValue, present := chartProperty.Tags().Lookup(featureCarrier.Symbol)
		if !present {
			continue
		}
		legendFeatureValue, check := legendOccurrence.Field(featureCarrier.Encounter.Property.FieldDescriptor)
		if !check.OK() {
			return fmt.Errorf("Feature %s destination: %s", featureCarrier.Symbol, check.Why())
		}
		legendFeatureAddress, check := legendFeatureValue.Address()
		if !check.OK() {
			return fmt.Errorf("Feature %s address: %s", featureCarrier.Symbol, check.Why())
		}
		if check := values.Fill(legendFeatureAddress, chartTagValue); !check.OK() {
			return fmt.Errorf("Feature %s: %s", featureCarrier.Symbol, check.Why())
		}
	}
	if err := captureBarrels(barrelCarriers, chartCarrierType, legendOccurrence); err != nil {
		return err
	}
	for _, qualifierCarrier := range qualifierCarriers {
		if !chartMark.HasQualifier(qualifierCarrier.Symbol) {
			continue
		}
		legendQualifierValue, check := legendOccurrence.Field(qualifierCarrier.Property.FieldDescriptor)
		if !check.OK() {
			return fmt.Errorf("Qualifier %s destination: %s", qualifierCarrier.Symbol, check.Why())
		}
		legendQualifierAddress, check := legendQualifierValue.Address()
		if !check.OK() {
			return fmt.Errorf("Qualifier %s address: %s", qualifierCarrier.Symbol, check.Why())
		}
		if check := values.Mark(legendQualifierAddress); !check.OK() {
			return fmt.Errorf("Qualifier %s presence: %s", qualifierCarrier.Symbol, check.Why())
		}
		if qualifierCarrier.Declaration != nil {
			legendQualifier, check := legendQualifierValue.AsStruct()
			if !check.OK() {
				return fmt.Errorf("Qualifier %s body: %s", qualifierCarrier.Symbol, check.Why())
			}
			if err := captureBarrels(qualifierCarrier.Declaration.Barrels, chartCarrierType, legendQualifier); err != nil {
				return fmt.Errorf("Qualifier %s: %w", qualifierCarrier.Symbol, err)
			}
		}
	}
	if spot.For == declarations.SpotForRegion {
		return p.scope(chartStruct, chartOccurrenceIndex, legendOccurrence, spot.Region.Declaration.Lanes)
	}
	return nil
}
