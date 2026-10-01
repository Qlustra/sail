package sail

import (
	"reflect"

	"github.com/qlustra/sail/internals/utilities"
)

type reflectable[T familyMarker] struct {
}

// ======================================== Directions

func (reflectable[T]) Target() reflect.Type {
	return reflect.TypeFor[T]()
}

// ======================================== Lane

type LaneExplorationStrategy string

const (
	ExploreEagerly        LaneExplorationStrategy = "eager"
	ExploreConservatively LaneExplorationStrategy = "conservative"
	ExploreFreely         LaneExplorationStrategy = "free"
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

type Spot[M familyMarker] struct {
	reflectable[M]
}

func (Spot[M]) spot() {}

type spotDeclaration interface {
	spot()
	Target() reflect.Type
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
type chart interface {
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
