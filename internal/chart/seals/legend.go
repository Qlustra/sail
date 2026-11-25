// Package seals defines membership-only interfaces. Their private methods are
// intentionally never invoked; concrete types embed them to opt into categories.
package seals

type Marker interface {
	marker()
}

type Legend interface {
	Marker
	legend()
}

type Lane interface {
	Marker
	lane()
}

type Region interface {
	Marker
	region()
}

type Landmark interface {
	Marker
	landmark()
}

type Feature interface {
	Marker
	feature()
}

type Barrel interface {
	Marker
	barrel()
}

type Qualifier interface {
	Marker
	qualifier()
}
