package sail

import (
	"fmt"
	"reflect"
	"slices"
)

// lanePopulation owns the counts and destinations for one populated location.
// Both strategies consume forward: exploration already proved unambiguous
// choices, while eager explicitly permits the current lane to take precedence.
type lanePopulation struct {
	location   *location
	value      reflect.Value
	families   markFamilies
	index      int
	counts     []int
	subjects   []reflect.Value
	candidates []candidate
}

func newLanePopulation(location *location, value reflect.Value, families markFamilies) (*lanePopulation, error) {
	p := &lanePopulation{location: location, value: value, families: families}
	return p, p.enter()
}

func (p *lanePopulation) enter() error {
	if p.index == len(p.location.lanes) {
		return nil
	}
	lane := p.location.lanes[p.index]
	value, err := declarationCarrier(p.value, lane.schemaField)
	if err != nil {
		return fmt.Errorf("%v.%s: %w", p.location.typ, p.location.laneName(p.index), err)
	}
	p.counts = make([]int, len(lane.spots))
	p.subjects = make([]reflect.Value, len(lane.spots))
	p.candidates = p.location.candidates(p.index)
	for i, spot := range lane.spots {
		p.subjects[i], err = declarationField(value, slices.Concat(spot.schemaField, spot.subjectField))
		if err != nil {
			return fmt.Errorf("%v.%s, spot %s: %w", p.location.typ, p.location.laneName(p.index), spot.target.mark, err)
		}
	}
	return nil
}

func (p *lanePopulation) selectField(field reflect.StructField) (*match, error) {
	for p.index < len(p.location.lanes) {
		selected, err := survey(field, p.candidates, p.families)
		if err != nil {
			return nil, err
		}
		if selected != nil {
			index := selected.candidate.spotIndex
			spot := p.location.lanes[p.index].spots[index]
			if spot.cardinality.max >= 0 && p.counts[index] >= spot.cardinality.max {
				return nil, fmt.Errorf("%v.%s, spot %s: exceeds maximum of %d occurrences", p.location.typ, p.location.laneName(p.index), spot.target.mark, spot.cardinality.max)
			}
			return selected, nil
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("mark %s:%q has no accepting spot in the remaining lanes of %v", p.location.mark.tag, field.Tag.Get(p.location.mark.tag), p.location.typ)
}

func (p *lanePopulation) store(index int, occurrence reflect.Value) {
	subject := p.subjects[index]
	switch p.location.lanes[p.index].spots[index].kind {
	case singleSpot:
		subject.Set(occurrence)
	case optionalSpot:
		subject.Set(occurrence.Addr())
	case multiSpot:
		subject.Set(reflect.Append(subject, occurrence))
	}
	p.counts[index]++
}

func (p *lanePopulation) advance() error {
	for index, spot := range p.location.lanes[p.index].spots {
		if p.counts[index] < spot.cardinality.min {
			return fmt.Errorf("%v.%s, spot %s: requires at least %d occurrences, got %d", p.location.typ, p.location.laneName(p.index), spot.target.mark, spot.cardinality.min, p.counts[index])
		}
	}
	p.index++
	return p.enter()
}

func (p *lanePopulation) finish() error {
	for p.index < len(p.location.lanes) {
		if err := p.advance(); err != nil {
			return err
		}
	}
	return nil
}
