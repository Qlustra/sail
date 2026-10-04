package legend

import "github.com/qlustra/sail/internal/declaration"

// Feature stores the text read from its named Application tag during population.
type Feature = declaration.Feature

type FeatureDeclaration interface {
	Symbol() string
}
