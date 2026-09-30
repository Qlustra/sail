package sail

import (
	"fmt"
	"slices"
	"strings"
)

// mark identifies one location within a tag's mark family. It owns its
// normalized discriminator slice; declarations may safely reuse their input.
type mark struct {
	tag           string
	family        string
	discriminants []string
}

type markFamily struct {
	tag    string
	family string
}

// Discriminator vocabulary is global to a family, even where only some of the
// family's locations are eligible. This prevents an unavailable variant from
// being interpreted as an ordinary mark with extra qualifiers.
type markFamilies map[markFamily]map[string]struct{}

func newMark(tag, family string, discriminants []string) (mark, error) {
	if err := validateTagKey(tag); err != nil {
		return mark{}, err
	}
	if err := validateTagComponent(family, true); err != nil {
		return mark{}, fmt.Errorf("invalid mark family: %w", err)
	}

	m := mark{
		tag:           tag,
		family:        family,
		discriminants: slices.Clone(discriminants),
	}
	slices.Sort(m.discriminants)
	for i, discriminator := range m.discriminants {
		if err := validateTagComponent(discriminator, false); err != nil {
			return mark{}, fmt.Errorf("invalid mark discriminator: %w", err)
		}
		if i > 0 && m.discriminants[i-1] == discriminator {
			return mark{}, fmt.Errorf("duplicate mark discriminator %q", discriminator)
		}
	}
	return m, nil
}

// compileMarkFamilies runs after graph discovery, when every reachable
// location has contributed its discriminator vocabulary.
func compileMarkFamilies(locations []*location) (markFamilies, error) {
	families := make(markFamilies)
	for _, location := range locations {
		if location == nil {
			return nil, fmt.Errorf("cannot compile a nil location")
		}
		family := markFamily{tag: location.mark.tag, family: location.mark.family}
		vocabulary := families[family]
		if vocabulary == nil {
			vocabulary = make(map[string]struct{})
			families[family] = vocabulary
		}
		for _, discriminator := range location.mark.discriminants {
			vocabulary[discriminator] = struct{}{}
		}
	}

	for _, location := range locations {
		family := markFamily{tag: location.mark.tag, family: location.mark.family}
		for _, qualifier := range location.qualifiers {
			if _, conflict := families[family][qualifier.symbol]; conflict {
				return nil, fmt.Errorf("location %v: qualifier %q conflicts with a discriminator in mark family %q/%q", location.typ, qualifier.symbol, family.tag, family.family)
			}
		}
	}
	for _, location := range locations {
		for laneIndex, lane := range location.lanes {
			type identity struct {
				tag, family, discriminants string
			}
			seen := make(map[identity]int)
			for spotIndex, spot := range lane.spots {
				if spot.target == nil {
					return nil, fmt.Errorf("location %v, lane %d, spot %d: missing target location", location.typ, laneIndex, spotIndex)
				}
				mark := spot.target.mark
				key := identity{tag: mark.tag, family: mark.family, discriminants: strings.Join(mark.discriminants, ",")}
				if previous, duplicate := seen[key]; duplicate {
					return nil, fmt.Errorf("location %v, lane %d: spots %d and %d have ambiguous identical marks %q/%q with discriminants %v", location.typ, laneIndex, previous, spotIndex, mark.tag, mark.family, mark.discriminants)
				}
				seen[key] = spotIndex
			}
		}
	}
	return families, nil
}
