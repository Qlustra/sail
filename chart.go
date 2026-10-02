package sail

import (
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

// ======================================== Lane

type LaneExplorationStrategy string

const (
	// ExploreEagerly stays in the current lane while it accepts a field.
	ExploreEagerly LaneExplorationStrategy = "eager"
	// ExploreUnambiguously requires structurally distinct lane choices. It is
	// the default when the exploration tag is omitted.
	ExploreUnambiguously LaneExplorationStrategy = "unambiguous"
)

type Lane struct {
}

func (Lane) lane() {}

type laneDeclaration interface {
	lane()
}

func isLaneField(t reflect.StructField) bool {
	return utilities.FieldImplements[laneDeclaration](t)
}

// ======================================== Spot

// SingleSpot describes exactly one occurrence of M, a non-pointer marker struct.
type SingleSpot[M familyMarker] struct {
	Subject M
}

// OptionalSpot describes zero or one occurrence of M. Subject is nil when absent;
// M itself must be a non-pointer marker struct.
type OptionalSpot[M familyMarker] struct {
	Subject *M
}

// MultiSpot describes occurrences of M, a non-pointer marker struct. Its field
// requires multiplicity:"any" for zero or more, or "some" for one or more.
type MultiSpot[M familyMarker] struct {
	Subject []M
}

type spotKind string

const (
	singleSpot   spotKind = "single"
	optionalSpot spotKind = "optional"
	multiSpot    spotKind = "multi"
)

// spotShape keeps the target tied to the type held by Subject. Its sealed
// provider is promoted by composite declarations without a public Target hook.
type spotShape struct {
	kind   spotKind
	target reflect.Type
}

func (SingleSpot[M]) spot() spotShape {
	return spotShape{kind: singleSpot, target: reflect.TypeFor[M]()}
}

func (OptionalSpot[M]) spot() spotShape {
	return spotShape{kind: optionalSpot, target: reflect.TypeFor[M]()}
}

func (MultiSpot[M]) spot() spotShape {
	return spotShape{kind: multiSpot, target: reflect.TypeFor[M]()}
}

type spotDeclaration interface {
	spot() spotShape
}

func isSpotField(t reflect.StructField) bool {
	return utilities.FieldImplements[spotDeclaration](t)
}

// ======================================== Marker

type marker struct {
}

// The method name must differ from the embedded marker field to be promoted.
func (marker) isMarker() {}

type MarkerDeclaration interface {
	isMarker()
}

// Hook requirements belong to API boundaries, independently of role recognition.
type chart interface { // legend
	chartDeclaration
	MarkProvider
}

type familyMarker interface {
	MarkerDeclaration
	FamilyProvider
}

type Chart struct {
	marker
}

func (Chart) chart() {}

type chartDeclaration interface {
	chart()
}

func isChart(f reflect.StructField) bool {
	return utilities.FieldImplements[chartDeclaration](f)
}

type Landmark struct {
	marker
}

func (Landmark) landmark() {}

type landmarkDeclaration interface {
	landmark()
}

func isLandmarkField(f reflect.StructField) bool {
	return utilities.FieldImplements[landmarkDeclaration](f)
}

type Region struct {
	marker
}

func (Region) region() {}

type regionDeclaration interface {
	region()
}

func isRegionField(f reflect.StructField) bool {
	return utilities.FieldImplements[regionDeclaration](f)
}

var markerDeclarationType = reflect.TypeOf((*MarkerDeclaration)(nil)).Elem()

// ======================================== Mark

// MarkProvider supplies the chart's tag namespace, inherited by its locations.
// Declaration hooks run on a newly allocated zero-valued declaration and must
// not require initialization, including when promoted through an embedded pointer.
type MarkProvider interface {
	Mark() string
}

// FamilyProvider supplies a region or landmark's family within the chart's tag.
// Family follows the same zero-value contract as Mark.
type FamilyProvider interface {
	Family() string
}

// DiscriminantProvider optionally distinguishes variants of the same family.
// Discriminate follows the same zero-value contract as Mark. Its result is copied
// before normalization, so declarations may safely reuse the returned slice.
type DiscriminantProvider interface {
	Discriminate() []string
}

// ======================================== Recognizer

// ======================================== Qualifier

type MarkQualifier struct {
	Symbol string
}

type Qualifier struct {
	present bool
}

func (Qualifier) qualifier() {}

type qualifierDeclaration interface {
	qualifier()
}

func (q *Qualifier) Present() bool {
	return q.present
}

func (q *Qualifier) Set() {
	q.present = true
}

func (q *Qualifier) Unset() {
	q.present = false
}

type QualifierDeclaration interface {
	Symbol() string
}

// ======================================== Feature

type Feature struct {
	value string
}

func (Feature) feature() {}

type featureDeclaration interface {
	feature()
}

func (f *Feature) Get() string {
	return f.value
}

func (f *Feature) Set(value string) {
	f.value = value
}

type FeatureDeclaration interface {
	Symbol() string
}

// ======================================== Capability

type Behavior []MarkerCapability

type MarkerCapability struct {
	typ reflect.Type
}

func (c MarkerCapability) Matches(t reflect.Type) bool {
	if t.Implements(c.typ) {
		return true
	}

	if t.Kind() != reflect.Pointer {
		return reflect.PointerTo(t).Implements(c.typ)
	}

	return false
}

// Capability records a type without validating it. Exploration requires every
// declared behavior capability to describe an interface.
func Capability[I any]() MarkerCapability {
	return MarkerCapability{typ: reflect.TypeFor[I]()}
}

// BehaviorProvider describes a declaration's capabilities. Behavior must work
// on a newly allocated zero-valued declaration without prior initialization.
type BehaviorProvider interface {
	Behavior() Behavior
}

var behaviorProviderType = reflect.TypeOf((*BehaviorProvider)(nil)).Elem()
