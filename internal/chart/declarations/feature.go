package declarations

import (
	"fmt"

	"github.com/qlustra/intel/logic"
	"github.com/qlustra/sail/internal/chart/carriers"
)

const (
	MissingFeatureCode         logic.CheckCode = "missing_feature"
	DuplicateFeatureSymbolCode logic.CheckCode = "duplicate_feature_symbol"
)

// addFeature owns declaration-local placement, not intrinsic symbol validation.
// Failed additions leave earlier work in place; compilation discards the session.
func addFeature(features []*carriers.Feature, feature *carriers.Feature) ([]*carriers.Feature, logic.Check) {
	if feature == nil {
		return features, logic.Nope(MissingFeatureCode, "cannot add a nil Feature")
	}
	for _, existing := range features {
		if existing != nil && existing.Symbol == feature.Symbol {
			return features, logic.Nope(DuplicateFeatureSymbolCode, fmt.Sprintf("symbol %q: %s conflicts with %s", feature.Symbol, feature, existing))
		}
	}
	return append(features, feature), logic.Ok()
}
