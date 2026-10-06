package engine

import "reflect"

// Chart holds the reusable descriptors compiled from a legend declaration.
// Populate uses it to describe an application declaration in a fresh world.
type Chart struct {
	Root  *markerDescriptor
	Codex Codex

	families markFamilies
}

type Codex struct {
	Markers map[reflect.Type]*markerDescriptor
}
