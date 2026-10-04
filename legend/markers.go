// Package legend provides the declaration vocabulary used to explore a Sail
// Chart and the typed storage populated in a World.
package legend

import "github.com/qlustra/sail/internal/declaration"

// Legend is embedded in the root declaration used to explore a Chart.
// Its populated value carries the World's reference to the Application root.
type Legend = declaration.Legend

// Region describes an Application occurrence with child lanes.
type Region = declaration.Region

// Landmark describes an Application occurrence without child lanes.
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
