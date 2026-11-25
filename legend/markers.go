package legend

import (
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/sail/internal/chart/behavior"
	"github.com/qlustra/sail/internal/chart/seals"
	"github.com/qlustra/sail/internal/chart/values"
)

// ================================================== Legend

// Legend marks the root grammar model. Embed it with sail:"legend" and a
// mark tag naming the grammar's Chart tag key.
type Legend struct {
	seals.Legend
	define.Scope
}

// ================================================== Lane

// Lane marks a body grouping Spots. A Legend or Region uses a Lane by
// declaring a field tagged sail:"lane".
type Lane struct {
	seals.Lane
	define.Scope
}

// ================================================== Region

type occurrenceState = values.Occurrence

// Region marks a grammar token that opens a nested scope. Embed it with
// sail:"region" and a family tag. A populated occurrence reports presence
// and records its Chart field location even when it contains no children.
type Region struct {
	occurrenceState
	seals.Region
	define.Scope
}

// ================================================== Landmark

// Landmark marks a grammar leaf. Embed it with sail:"landmark" and a family
// tag. A populated occurrence reports presence and records its Chart field
// location. Landmarks may carry descriptive values but cannot open scopes.
type Landmark struct {
	occurrenceState
	seals.Landmark
	define.Leaf
}

// ================================================== Feature

type featureValue = values.Value[string]

// Feature reads an occurrence's tag value; Present distinguishes an empty value from absence.
type Feature struct {
	featureValue
	seals.Feature
	define.Leaf
}

// ================================================== Barrel

type valueCell[V any] = values.Value[V]
type barrelInspection = values.Inspection
type barrelTypes[V, I any] = behavior.BarrelTypes[V, I]

// Barrel captures a provider method's result on a fresh receiver initialized
// from the Chart carrier type's zero value, not on a live Chart field. Declare it with
// sail:"barrel" and a populate tag naming an exported method with no arguments
// and one result assignable to V.
//
// I is the provider interface. An empty interface defers method resolution to
// population. An unavailable provider leaves an optional Barrel absent;
// barrel:"required" makes unavailability an error when the Barrel is active.
type Barrel[V, I any] struct {
	barrelTypes[V, I]
	valueCell[V]
	barrelInspection
	seals.Barrel
	define.Leaf
}

// Get returns the captured value assigned to V, or the zero value of V if no
// capture occurred. Use [Barrel.Present] to tell those cases apart.
func (b Barrel[V, I]) Get() V { return b.valueCell.Get() }

// Present reports whether a value was captured, including a zero or nil result.
func (b Barrel[V, I]) Present() bool { return b.valueCell.Present() }

// Method returns metadata for the provider method used for capture. It is
// unbound to any receiver. If no capture occurred, the check fails.
func (b Barrel[V, I]) Method() (*capture.MethodInterrogator, logic.Check) {
	return b.barrelInspection.Method()
}

// Result returns the provider's declared result type and captured value as an
// Intel interrogator. The declared type can differ from V, the assigned view
// returned by [Barrel.Get]. If no capture occurred, the check fails.
func (b Barrel[V, I]) Result() (*capture.ResultInterrogator, logic.Check) {
	return b.barrelInspection.Result()
}

// CarrierType returns an interrogator for the original Chart property type,
// not the surrogate receiver type or the declared result type. If no capture
// occurred, the check fails.
func (b Barrel[V, I]) CarrierType() (*capture.TypeInterrogator, logic.Check) {
	return b.barrelInspection.CarrierType()
}

// ================================================== Qualifier

type qualifierPresence = values.Presence

// Qualifier reports token presence without assigning it domain meaning.
type Qualifier struct {
	qualifierPresence
	seals.Qualifier
	define.Leaf
}

// Intel matches field value types, so verify the value method sets rather than
// only their pointers.
var (
	_ behavior.Legend    = Legend{}
	_ behavior.Lane      = Lane{}
	_ behavior.Region    = Region{}
	_ behavior.Landmark  = Landmark{}
	_ behavior.Feature   = Feature{}
	_ behavior.Barrel    = Barrel[any, any]{}
	_ behavior.Qualifier = Qualifier{}
)
