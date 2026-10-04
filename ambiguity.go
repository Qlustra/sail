package sail

import "fmt"

// laneMarks records spot identities for collision checks. Map iteration never
// selects a candidate or determines which ambiguity is reported first.
type laneMarks map[markIdentity]int

func (d laneDescriptor) markSet() (laneMarks, error) {
	marks := make(laneMarks, len(d.spots))
	for index, spot := range d.spots {
		if spot.target == nil {
			return nil, fmt.Errorf("spot %d: missing target descriptor", index)
		}
		mark := spot.target.mark
		identity := mark.identity()
		if previous, duplicate := marks[identity]; duplicate {
			return nil, fmt.Errorf("spots %d and %d have ambiguous identical mark %s", previous, index, mark)
		}
		marks[identity] = index
	}
	return marks, nil
}

func (d laneDescriptor) overlap(marks laneMarks) (mark, bool) {
	for _, spot := range d.spots {
		mark := spot.target.mark
		if _, found := marks[mark.identity()]; found {
			return mark, true
		}
	}
	return mark{}, false
}

// validateLaneAmbiguity runs on the completed graph before it is published.
// Every lane rejects ambiguous spots, regardless of its exploration strategy.
func validateLaneAmbiguity(descriptors []*markerDescriptor) error {
	for _, descriptor := range descriptors {
		if err := descriptor.validateLanes(); err != nil {
			return err
		}
	}
	return nil
}

func (d *markerDescriptor) validateLanes() error {
	marks := make([]laneMarks, len(d.lanes))
	for index, lane := range d.lanes {
		set, err := lane.markSet()
		if err != nil {
			return fmt.Errorf("%v.%s: %w", d.typ, d.laneName(index), err)
		}
		marks[index] = set
	}
	for index, lane := range d.lanes {
		if lane.strategy == ExploreUnambiguously {
			if err := d.validateFollowingLanes(index, marks); err != nil {
				return err
			}
		}
	}
	return nil
}

// A nullable lane may disappear, exposing the following lane. Include the first
// required lane in lookahead, but do not cross it. Finite maximum counts do not
// resolve structural overlap; occurrence counts are population constraints.
func (d *markerDescriptor) validateFollowingLanes(index int, marks []laneMarks) error {
	current := d.lanes[index]
	for next := index + 1; next < len(d.lanes); next++ {
		if mark, overlaps := current.overlap(marks[next]); overlaps {
			return fmt.Errorf("%v.%s: unambiguous exploration overlaps reachable lane %s on mark %s", d.typ, d.laneName(index), d.laneName(next), mark)
		}
		if !d.lanes[next].nullable() {
			break
		}
	}
	return nil
}

func (d *markerDescriptor) laneName(index int) string {
	return d.typ.FieldByIndex(d.lanes[index].declarationPath).Name
}
