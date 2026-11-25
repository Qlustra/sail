package declarations

import (
	"fmt"

	"github.com/qlustra/intel/logic"
)

// ValidateQualifierDiscriminators rejects tokens used as both a discriminator and
// Qualifier symbol in one grammar family, independent of Lane or nested scope.
func ValidateQualifierDiscriminators(legend *Legend) logic.Check {
	_, check := BuildVocabulary(legend)
	return check
}

// BuildVocabulary indexes the completed graph without modifying it.
func BuildVocabulary(legend *Legend) (*Vocabulary, logic.Check) {
	if legend == nil {
		return nil, logic.Nope("invalid_qualifier_discriminators", "legend declaration is nil")
	}
	vocabulary := &Vocabulary{families: make(map[string]map[string]tokenRole)}
	type key struct{ family, token string }
	type origin struct{ kind, use, field string }
	discriminators := make(map[key]origin)
	qualifiers := make(map[key]origin)
	conflict := func(k key, q, d origin) logic.Check {
		return logic.Nope("qualifier_discriminator_collision", fmt.Sprintf("family=%q token=%q: qualifier field %q in %s use %q conflicts with discriminator in %s use %q", k.family, k.token, q.field, q.kind, q.use, d.kind, d.use))
	}
	check := walkBodies(legend, nil, func(spot *Spot) logic.Check {
		var family string
		var tokens []string
		var qs []*QualifierCarrier
		kind := string(spot.For)
		if spot.For == SpotForRegion {
			r := spot.Region.Declaration
			family, tokens, qs = r.Family, r.Discriminators, r.Qualifiers
		} else {
			l := spot.Landmark.Declaration
			family, tokens, qs = l.Family, l.Discriminators, l.Qualifiers
		}
		roles := vocabulary.families[family]
		if roles == nil {
			roles = make(map[string]tokenRole)
			vocabulary.families[family] = roles
		}
		for _, token := range tokens {
			k := key{family, token}
			d := origin{kind: kind, use: spotField(spot)}
			if q, ok := qualifiers[k]; ok {
				return conflict(k, q, d)
			}
			if _, ok := discriminators[k]; !ok {
				discriminators[k] = d
			}
			roles[token] = discriminatorToken
		}
		for _, qualifier := range qs {
			if qualifier == nil {
				continue
			}
			field := "<unknown>"
			if qualifier.Property != nil {
				field = qualifier.Property.FieldName()
			}
			k := key{family, qualifier.Symbol}
			q := origin{kind: kind, use: spotField(spot), field: field}
			if d, ok := discriminators[k]; ok {
				return conflict(k, q, d)
			}
			if _, ok := qualifiers[k]; !ok {
				qualifiers[k] = q
			}
			roles[qualifier.Symbol] = qualifierToken
		}
		return logic.Ok()
	})
	if !check.OK() {
		return nil, check
	}
	return vocabulary, check
}
