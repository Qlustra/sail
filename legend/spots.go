package legend

import "github.com/qlustra/sail/internal/declaration"

type familyMarker interface {
	MarkerDeclaration
	FamilyProvider
}

// SingleSpot describes exactly one occurrence of M, a non-pointer marker struct.
type SingleSpot[M familyMarker] = declaration.SingleSpot[M]

// OptionalSpot describes zero or one occurrence of M. Value is nil when absent;
// M itself must be a non-pointer marker struct.
type OptionalSpot[M familyMarker] = declaration.OptionalSpot[M]

// MultiSpot describes occurrences of M, a non-pointer marker struct. Its field
// requires multiplicity:"any" for zero or more, or "some" for one or more.
type MultiSpot[M familyMarker] = declaration.MultiSpot[M]
