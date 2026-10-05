package engine

import (
	"fmt"
	"reflect"
	"slices"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// Occurrence retains one populated marker and its position in the World.
// Descriptors are shared; occurrence identity belongs to existing World storage.
type Occurrence struct {
	Value      reflect.Value
	Source     decl.Reference
	Parent     *Occurrence
	LanePath   []int
	SpotPath   []int
	ItemIndex  int
	descriptor *markerDescriptor
}

func navigationRoot(chart *Chart, world reflect.Value) (*Occurrence, error) {
	if chart == nil || chart.Root == nil {
		return nil, fmt.Errorf("sail: traversal requires an explored chart")
	}
	root, err := markerOccurrence(chart.Root, world)
	if err != nil {
		return nil, fmt.Errorf("sail: traversal world: %w", err)
	}
	root.ItemIndex = -1
	return root, nil
}

func markerOccurrence(descriptor *markerDescriptor, value reflect.Value) (*Occurrence, error) {
	value, err := reflectutils.Dereference(value)
	if err != nil {
		return nil, err
	}
	if value.Type() != descriptor.typ {
		return nil, fmt.Errorf("marker type %v does not match chart declaration %v", value.Type(), descriptor.typ)
	}
	if !value.CanAddr() || !value.CanInterface() {
		return nil, fmt.Errorf("marker %v requires addressable World storage", descriptor.typ)
	}
	carrier, err := reflectutils.FieldByPath(value, descriptor.carrierPath)
	if err == nil {
		carrier, err = reflectutils.Dereference(carrier)
	}
	if err != nil {
		return nil, fmt.Errorf("marker %v: %w", descriptor.typ, err)
	}
	source := decl.ReferenceOf(carrier.Interface())
	if source.Root == nil || source.Type == nil {
		return nil, fmt.Errorf("marker %v has not been populated", descriptor.typ)
	}
	source.Path = slices.Clone(source.Path)
	return &Occurrence{Value: value, Source: source, descriptor: descriptor}, nil
}

// immediateChildren follows only lane and spot descriptors. It neither reads
// live Application storage nor allocates missing World carriers.
func immediateChildren(parent *Occurrence) ([]*Occurrence, error) {
	var children []*Occurrence
	for _, lane := range parent.descriptor.lanes {
		members, err := laneOccurrences(parent, lane)
		if err != nil {
			return nil, fmt.Errorf("%s, lane %v: %w", parent.Source, lane.declarationPath, err)
		}
		// Spots are alternatives, not an ordering policy. Restore the source
		// declaration order within each lane before exposing its occurrences.
		slices.SortStableFunc(members, func(a, b *Occurrence) int {
			return slices.Compare(a.Source.Path, b.Source.Path)
		})
		children = append(children, members...)
	}
	return children, nil
}

func laneOccurrences(parent *Occurrence, lane laneDescriptor) ([]*Occurrence, error) {
	value, err := reflectutils.FieldByPath(parent.Value, lane.declarationPath)
	if err != nil {
		return nil, err
	}
	var children []*Occurrence
	for _, spot := range lane.spots {
		members, err := spotOccurrences(value, spot)
		if err != nil {
			return nil, fmt.Errorf("spot %v: %w", spot.declarationPath, err)
		}
		for index, member := range members {
			member.Parent = parent
			member.LanePath = slices.Clone(lane.declarationPath)
			member.SpotPath = slices.Clone(spot.declarationPath)
			member.ItemIndex = index
		}
		children = append(children, members...)
	}
	return children, nil
}

func spotOccurrences(lane reflect.Value, spot spotDescriptor) ([]*Occurrence, error) {
	value, err := reflectutils.FieldByPath(lane, slices.Concat(spot.declarationPath, spot.valuePath))
	if err != nil {
		return nil, err
	}
	if spot.kind == decl.OptionalSpotKind && value.IsNil() {
		return nil, nil
	}
	if spot.kind != decl.MultiSpotKind {
		member, err := markerOccurrence(spot.target, value)
		if err != nil {
			return nil, err
		}
		return []*Occurrence{member}, nil
	}
	members := make([]*Occurrence, 0, value.Len())
	for index := 0; index < value.Len(); index++ {
		member, err := markerOccurrence(spot.target, value.Index(index))
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", index, err)
		}
		members = append(members, member)
	}
	return members, nil
}
