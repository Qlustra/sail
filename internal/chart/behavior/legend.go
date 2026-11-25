// Package behavior defines the contracts consumed by the chart compiler,
// combining membership seals with the APIs that concrete types implement.
package behavior

import (
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/seals"
)

type Marker interface {
	seals.Marker
}

type Legend interface {
	seals.Legend
	define.ScopeIndicator
}

type Lane interface {
	seals.Lane
	define.ScopeIndicator
}

type Region interface {
	Marker
	seals.Region
	define.ScopeIndicator
}

// An outer authored Landmark may implement Accept() typerule.RuleAPI. The
// retained instruction is evaluated on the original consumer type after selection;
// nil is a binding error, while rule semantics are checked with a real subject.
// Regions establish scopes and do not have this hook.
type Landmark interface {
	Marker
	seals.Landmark
	define.LeafIndicator
}

type Feature interface {
	Marker
	seals.Feature
	define.LeafIndicator
}

type Barrel interface {
	Marker
	seals.Barrel
	define.LeafIndicator
	barrelShape() ValueShape
}

type Qualifier interface {
	Marker
	seals.Qualifier
	define.LeafIndicator
}

// ValueShape describes the fixed stored result and consumer capability used by
// declaration-time value bindings.
type ValueShape struct {
	Result     *capture.TypeInterrogator
	Capability *capture.TypeInterrogator
}

// BarrelTypes carries a Barrel's type arguments into the compiler without
// adding a public method to the marker or mixing type metadata with capture state.
type BarrelTypes[V, I any] struct{}

func (BarrelTypes[V, I]) barrelShape() ValueShape {
	return ValueShape{Result: capture.Type[V](), Capability: capture.Type[I]()}
}

// BarrelShape reads the declaration-time shape through the private marker protocol.
func BarrelShape(barrel Barrel) ValueShape { return barrel.barrelShape() }
