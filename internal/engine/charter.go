package engine

import "reflect"

// Charter holds the reusable descriptors compiled from a legend declaration.
// Populate uses it to describe a subject declaration in a fresh chart.
type Charter struct {
	Root  *markerDescriptor
	Codex Codex

	families markFamilies
}

type Codex struct {
	Markers map[reflect.Type]*markerDescriptor
}
