package sail

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
)

// Course will describe a live application traversal.
type Course struct{}

// ErrTraversalNotImplemented indicates that live traversal and hydration are
// not available yet. Chart compilation and field surveying are implemented.
var ErrTraversalNotImplemented = errors.New("sail: live application traversal is not implemented")

func (c *Chart[L]) Traverse(app any, course *Course) error {
	return ErrTraversalNotImplemented
}

// A candidate preserves its occurrence in the enclosing descriptor's lanes
// and spots, while sharing the target's reusable descriptor.
type candidate struct {
	laneIndex int
	spotIndex int
	target    *markerDescriptor
}

type match struct {
	candidate  candidate
	qualifiers []string
}

// survey selects a descriptor by its compound mark among the caller's eligible
// candidates. Lane advancement and receiving-type validation belong to traversal.
// A nil match without an error means that no eligible candidate matched.
func survey(field reflect.StructField, candidates []candidate, families markFamilies) (*match, error) {
	parsedTags := make(map[string]parsedMarkTag)
	missingTags := make(map[string]bool)
	var matches []match

	for _, candidate := range candidates {
		if candidate.target == nil {
			return nil, fmt.Errorf("field %s: candidate has no target descriptor", field.Name)
		}
		mark := candidate.target.mark
		if missingTags[mark.tag] {
			continue
		}
		parsed, cached := parsedTags[mark.tag]
		if !cached {
			raw, present := field.Tag.Lookup(mark.tag)
			if !present {
				missingTags[mark.tag] = true
				continue
			}
			var err error
			parsed, err = parseMarkTag(raw)
			if err != nil {
				return nil, fmt.Errorf("field %s, tag %q: %w", field.Name, mark.tag, err)
			}
			parsedTags[mark.tag] = parsed
		}
		if parsed.family != mark.family {
			continue
		}

		vocabulary, present := families[markFamily{tag: mark.tag, family: mark.family}]
		if !present {
			return nil, fmt.Errorf("field %s: uncompiled mark family %q/%q", field.Name, mark.tag, mark.family)
		}
		var discriminants, qualifiers []string
		for _, value := range parsed.values {
			if _, discriminator := vocabulary[value]; discriminator {
				discriminants = append(discriminants, value)
			} else {
				qualifiers = append(qualifiers, value)
			}
		}
		slices.Sort(discriminants)
		if slices.Equal(discriminants, mark.discriminants) {
			matches = append(matches, match{candidate: candidate, qualifiers: qualifiers})
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("field %s: ambiguous mark matches %d candidate occurrences", field.Name, len(matches))
	}
	selected := &matches[0]
	for _, value := range selected.qualifiers {
		declared := slices.ContainsFunc(selected.candidate.target.qualifiers, func(q qualifierDescriptor) bool { return q.symbol == value })
		if !declared {
			return nil, fmt.Errorf("field %s: unknown qualifier %q for marker %v", field.Name, value, selected.candidate.target.typ)
		}
	}
	return selected, nil
}
