package declarations

import (
	"fmt"
	"sort"
	"strings"

	"github.com/qlustra/intel/logic"
)

type tokenRole uint8

const (
	discriminatorToken tokenRole = iota + 1
	qualifierToken
)

// Vocabulary is a read-only family-wide index built from the completed graph.
type Vocabulary struct {
	families map[string]map[string]tokenRole
}

// ChartMark contains only normalized, occurrence-local selector and qualifier data.
type ChartMark struct {
	selector   selector
	qualifiers []string
}

// Classify separates post-family modifiers by their compiled family-wide roles.
// It preserves exact spelling and never mutates the caller's modifier slice.
func (v *Vocabulary) Classify(family string, modifiers []string) (*ChartMark, logic.Check) {
	if v == nil {
		return nil, logic.Nope("invalid_recipe_token", "cannot classify with nil vocabulary")
	}
	roles, known := v.families[family]
	if !known {
		return nil, logic.Nope("invalid_recipe_token", fmt.Sprintf("unknown family %q", family))
	}
	seen := make(map[string]bool, len(modifiers))
	var discriminators, qualifiers []string
	for _, token := range modifiers {
		if token == "" {
			return nil, logic.Nope("invalid_recipe_token", fmt.Sprintf("family %q: empty modifier", family))
		}
		if seen[token] {
			return nil, logic.Nope("invalid_recipe_token", fmt.Sprintf("family %q: duplicate modifier %q", family, token))
		}
		if roles[token] == 0 {
			return nil, logic.Nope("invalid_recipe_token", fmt.Sprintf("family %q: unknown modifier %q", family, token))
		}
		seen[token] = true
		if roles[token] == discriminatorToken {
			discriminators = append(discriminators, token)
		} else {
			qualifiers = append(qualifiers, token)
		}
	}
	sort.Strings(discriminators)
	sort.Strings(qualifiers)
	return &ChartMark{selector: selector{family, strings.Join(discriminators, ",")}, qualifiers: qualifiers}, logic.Ok()
}

func (m *ChartMark) HasQualifier(symbol string) bool {
	if m == nil {
		return false
	}
	i := sort.SearchStrings(m.qualifiers, symbol)
	return i < len(m.qualifiers) && m.qualifiers[i] == symbol
}

// ValidateFor checks Qualifier applicability after exact Spot selection, without
// retrying selection or executing providers.
func (m *ChartMark) ValidateFor(spot *Spot) logic.Check {
	if m == nil || spot == nil {
		return logic.Nope("invalid_recipe_token", "cannot validate nil Chart mark or Spot")
	}
	var family string
	var qualifiers []*QualifierCarrier
	switch {
	case spot.For == SpotForRegion && spot.Region != nil && spot.Region.Declaration != nil:
		family, qualifiers = spot.Region.Declaration.Family, spot.Region.Declaration.Qualifiers
	case spot.For == SpotForLandmark && spot.Landmark != nil && spot.Landmark.Declaration != nil:
		family, qualifiers = spot.Landmark.Declaration.Family, spot.Landmark.Declaration.Qualifiers
	default:
		return logic.Nope("invalid_recipe_token", "cannot validate against invalid spot")
	}
	if family != m.selector.family {
		return logic.Nope("invalid_recipe_token", fmt.Sprintf("family %q does not match selected field %q family %q", m.selector.family, spotField(spot), family))
	}
	for _, token := range m.qualifiers {
		found := false
		for _, qualifier := range qualifiers {
			if qualifier != nil && qualifier.Symbol == token {
				found = true
				break
			}
		}
		if !found {
			return logic.Nope("invalid_recipe_token", fmt.Sprintf("family %q selected field %q does not support qualifier %q", family, spotField(spot), token))
		}
	}
	return logic.Ok()
}

func (m *ChartMark) String() string {
	if m == nil {
		return "<nil>"
	}
	return fmt.Sprintf("family=%q discriminate=%q qualifiers=%q", m.selector.family, m.selector.discriminators, m.qualifiers)
}
