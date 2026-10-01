package sail

import "fmt"

// laneMarks records spot identities for collision checks. Map iteration never
// selects a candidate or determines which ambiguity is reported first.
type laneMarks map[markIdentity]int

func (l lane) markSet() (laneMarks, error) {
	marks := make(laneMarks, len(l.spots))
	for index, spot := range l.spots {
		if spot.target == nil {
			return nil, fmt.Errorf("spot %d: missing target location", index)
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

func (l lane) overlap(marks laneMarks) (mark, bool) {
	for _, spot := range l.spots {
		mark := spot.target.mark
		if _, found := marks[mark.identity()]; found {
			return mark, true
		}
	}
	return mark{}, false
}

// validateLaneAmbiguity runs on the completed graph before it is published.
// Every lane rejects ambiguous spots, regardless of its exploration strategy.
func validateLaneAmbiguity(locations []*location) error {
	for _, location := range locations {
		if err := location.validateLanes(); err != nil {
			return err
		}
	}
	return nil
}

func (l *location) validateLanes() error {
	marks := make([]laneMarks, len(l.lanes))
	for index, lane := range l.lanes {
		set, err := lane.markSet()
		if err != nil {
			return fmt.Errorf("%v.%s: %w", l.typ, l.laneName(index), err)
		}
		marks[index] = set
	}
	for index, lane := range l.lanes {
		if lane.strategy == ExploreUnambiguously {
			if err := l.validateFollowingLanes(index, marks); err != nil {
				return err
			}
		}
	}
	return nil
}

// A nullable lane may disappear, exposing the following lane. Include the first
// required lane in lookahead, but do not cross it. Finite maximum counts do not
// resolve structural overlap; occurrence counts are population constraints.
func (l *location) validateFollowingLanes(index int, marks []laneMarks) error {
	current := l.lanes[index]
	for next := index + 1; next < len(l.lanes); next++ {
		if mark, overlaps := current.overlap(marks[next]); overlaps {
			return fmt.Errorf("%v.%s: unambiguous exploration overlaps reachable lane %s on mark %s", l.typ, l.laneName(index), l.laneName(next), mark)
		}
		if !l.lanes[next].nullable() {
			break
		}
	}
	return nil
}

func (l *location) laneName(index int) string {
	return l.typ.FieldByIndex(l.lanes[index].schemaField).Name
}
