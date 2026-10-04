package sail

import "reflect"

// Chart holds the reusable descriptors explored with legend L.
// Populate uses it to describe an application declaration in a fresh world.
type Chart[L legend] struct {
	Root  *markerDescriptor
	Codex Codex

	families markFamilies
}

type Codex struct {
	Markers map[reflect.Type]*markerDescriptor
}
