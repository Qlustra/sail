// Package declaration owns Legend carriers and their sealed engine protocols.
package declaration

type marker struct {
	source Reference
}

func (m *marker) setSource(source Reference) { m.source = source }

// The method name differs from the embedded marker field so it is promoted.
func (marker) isMarker() {}

type MarkerDeclaration interface {
	isMarker()
}

// Role recognition is independent of the hooks required at API boundaries.
type LegendRole interface {
	legend()
}

type LegendDeclaration interface {
	LegendRole
	MarkProvider
}

type FamilyMarker interface {
	MarkerDeclaration
	FamilyProvider
}

type Legend struct {
	marker
}

func (Legend) legend() {}

type Landmark struct {
	marker
}

func (Landmark) landmark() {}

type LandmarkRole interface {
	landmark()
}

type Region struct {
	marker
}

func (Region) region() {}

type RegionRole interface {
	region()
}

type MarkProvider interface {
	Mark() string
}

type FamilyProvider interface {
	Family() string
}

type DiscriminantProvider interface {
	Discriminate() []string
}
