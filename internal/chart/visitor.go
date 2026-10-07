package chart

import "fmt"

// Visitor processes one selected occurrence. Returning an error ends execution.
// Use On to receive a typed marker alongside its encounter context.
type Visitor func(Encounter) error

// Visitors processes each encounter in registration order. All matching typed
// visitors participate; the first error ends both the list and the traversal.
type Visitors []Visitor

// Visit creates a single-item visitor list from a typed callback.
func Visit[T any](visit func(T, Encounter) error) Visitors {
	return Visitors{On(visit)}
}

// On adapts a typed callback to one visitor for use in a Visitors list.
// Unmatched markers do not call visit; they remain part of the route.
// Pointer callbacks receive original Chart values. Interface callbacks prefer
// the marker's pointer over its value.
func On[T any](visit func(T, Encounter) error) Visitor {
	if visit == nil {
		return nil
	}
	return func(encounter Encounter) error {
		if encounter.occurrence == nil {
			return fmt.Errorf("sail: visitor requires a Chart encounter")
		}
		marker, matches := markerAs[T](encounter.occurrence.Value)
		if !matches {
			return nil
		}
		return visit(marker, encounter)
	}
}

func (visitors Visitors) validate() error {
	if len(visitors) == 0 {
		return fmt.Errorf("requires at least one visitor")
	}
	for index, visit := range visitors {
		if visit == nil {
			return fmt.Errorf("visitor %d is nil", index+1)
		}
	}
	return nil
}

func (visitors Visitors) visit(encounter Encounter) error {
	for index, visit := range visitors {
		if err := visit(encounter); err != nil {
			return fmt.Errorf("visitor %d: %w", index+1, err)
		}
	}
	return nil
}
