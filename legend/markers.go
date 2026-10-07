// Package legend provides the declaration vocabulary used to compile a Sail
// Charter and the typed storage populated in a Chart.
package legend

import "github.com/qlustra/sail/internal/declaration"

// Legend is embedded in the root declaration used to compile a Charter.
// Its populated value carries the Chart's reference to the Subject root.
type Legend = declaration.Legend

// Region describes a Subject occurrence with child lanes.
type Region = declaration.Region

// Landmark describes a Subject occurrence without child lanes.
type Landmark = declaration.Landmark

// Declaration is a Legend with its required Mark hook.
type Declaration = declaration.LegendDeclaration

// MarkerDeclaration identifies a Legend, Region, or Landmark carrier.
type MarkerDeclaration = declaration.MarkerDeclaration

// MarkProvider supplies the Legend's tag namespace, inherited by its descriptors.
// Declaration hooks run on newly allocated zero-valued declarations and must
// not require initialization, including when promoted through an embedded pointer.
type MarkProvider = declaration.MarkProvider

// FamilyProvider supplies a Region or Landmark's family within the Legend's tag.
// Family follows the same zero-value contract as Mark.
type FamilyProvider = declaration.FamilyProvider

// DiscriminantProvider optionally distinguishes variants of the same family.
// Discriminate follows the same zero-value contract as Mark. Sail copies its
// result before normalization, so declarations may safely reuse the returned slice.
type DiscriminantProvider = declaration.DiscriminantProvider
