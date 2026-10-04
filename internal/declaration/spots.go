package declaration

import "reflect"

type SingleSpot[M FamilyMarker] struct {
	Value M
}

type OptionalSpot[M FamilyMarker] struct {
	Value *M
}

type MultiSpot[M FamilyMarker] struct {
	Value []M
}

type SpotKind string

const (
	SingleSpotKind   SpotKind = "single"
	OptionalSpotKind SpotKind = "optional"
	MultiSpotKind    SpotKind = "multi"
)

// SpotShape keeps the target tied to the type held by Value. The sealed
// provider is promoted by composite declarations without a public Target hook.
type SpotShape struct {
	Kind    SpotKind
	Target  reflect.Type
	Carrier reflect.Type
}

func (SingleSpot[M]) spot() SpotShape {
	return SpotShape{Kind: SingleSpotKind, Target: reflect.TypeFor[M](), Carrier: reflect.TypeFor[SingleSpot[M]]()}
}

func (OptionalSpot[M]) spot() SpotShape {
	return SpotShape{Kind: OptionalSpotKind, Target: reflect.TypeFor[M](), Carrier: reflect.TypeFor[OptionalSpot[M]]()}
}

func (MultiSpot[M]) spot() SpotShape {
	return SpotShape{Kind: MultiSpotKind, Target: reflect.TypeFor[M](), Carrier: reflect.TypeFor[MultiSpot[M]]()}
}

type SpotDeclaration interface {
	spot() SpotShape
}

func DescribeSpot(provider SpotDeclaration) SpotShape { return provider.spot() }
