package engine

import (
	"fmt"
	"reflect"
	"slices"

	decl "github.com/qlustra/sail/internal/declaration"
)

// lanePopulation owns the counts and destinations for one populated marker.
// Both strategies consume forward: exploration already proved unambiguous
// choices, while eager explicitly permits the current lane to take precedence.
type lanePopulation struct {
	descriptor   *markerDescriptor
	value        reflect.Value
	families     markFamilies
	index        int
	counts       []int
	destinations []reflect.Value
	candidates   []candidate
}

func newLanePopulation(descriptor *markerDescriptor, value reflect.Value, families markFamilies) (*lanePopulation, error) {
	p := &lanePopulation{descriptor: descriptor, value: value, families: families}
	return p, p.enter()
}

func (p *lanePopulation) enter() error {
	if p.index == len(p.descriptor.lanes) {
		return nil
	}
	lane := p.descriptor.lanes[p.index]
	value, err := declarationCarrier(p.value, lane.declarationPath)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", p.descriptor.typ, p.descriptor.laneName(p.index), err)
	}
	p.counts = make([]int, len(lane.spots))
	p.destinations = make([]reflect.Value, len(lane.spots))
	p.candidates = p.descriptor.candidates(p.index)
	for i, spot := range lane.spots {
		p.destinations[i], err = declarationField(value, slices.Concat(spot.declarationPath, spot.valuePath))
		if err != nil {
			return fmt.Errorf("%v.%s, spot %s: %w", p.descriptor.typ, p.descriptor.laneName(p.index), spot.target.mark, err)
		}
	}
	return nil
}

func (p *lanePopulation) selectField(field reflect.StructField) (*match, error) {
	for p.index < len(p.descriptor.lanes) {
		selected, err := survey(field, p.candidates, p.families)
		if err != nil {
			return nil, err
		}
		if selected != nil {
			index := selected.candidate.spotIndex
			spot := p.descriptor.lanes[p.index].spots[index]
			if spot.cardinality.max >= 0 && p.counts[index] >= spot.cardinality.max {
				return nil, fmt.Errorf("%v.%s, spot %s: exceeds maximum of %d occurrences", p.descriptor.typ, p.descriptor.laneName(p.index), spot.target.mark, spot.cardinality.max)
			}
			return selected, nil
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("mark %s:%q has no accepting spot in the remaining lanes of %v", p.descriptor.mark.tag, field.Tag.Get(p.descriptor.mark.tag), p.descriptor.typ)
}

func (p *lanePopulation) store(index int, marker reflect.Value) {
	destination := p.destinations[index]
	switch p.descriptor.lanes[p.index].spots[index].kind {
	case decl.SingleSpotKind:
		destination.Set(marker)
	case decl.OptionalSpotKind:
		destination.Set(marker.Addr())
	case decl.MultiSpotKind:
		destination.Set(reflect.Append(destination, marker))
	}
	p.counts[index]++
}

func (p *lanePopulation) advance() error {
	for index, spot := range p.descriptor.lanes[p.index].spots {
		if p.counts[index] < spot.cardinality.min {
			return fmt.Errorf("%v.%s, spot %s: requires at least %d occurrences, got %d", p.descriptor.typ, p.descriptor.laneName(p.index), spot.target.mark, spot.cardinality.min, p.counts[index])
		}
	}
	p.index++
	return p.enter()
}

func (p *lanePopulation) finish() error {
	for p.index < len(p.descriptor.lanes) {
		if err := p.advance(); err != nil {
			return err
		}
	}
	return nil
}
