package sail

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internals/utilities"
)

type spotDescriptor struct {
	declarationPath []int
	valuePath       []int // Path inside the spot declaration to its carrier's Value.
	kind            spotKind
	cardinality     cardinality
	target          *markerDescriptor
}

// ======================================== Exploration

// Explore builds a chart with legend L, without an application declaration or instance.
func Explore[L legend]() (*Chart[L], error) {
	chart := &Chart[L]{}
	err := chart.Explore()
	return chart, err
}

// Explore publishes a complete chart only after all descriptors and mark
// families have been validated. Failed compilation leaves the chart unchanged.
func (c *Chart[L]) Explore() error {
	t := utilities.ResolvePointerChain(reflect.TypeFor[L]())
	if t == nil || t.Kind() != reflect.Struct || !utilities.TypeImplements[legendDeclaration](t) {
		return fmt.Errorf("%v is not a legend declaration", reflect.TypeFor[L]())
	}

	exp := explorer{
		codex:   Codex{Markers: make(map[reflect.Type]*markerDescriptor)},
		journal: NewJournal(),
	}
	root, err := exp.explore(t)
	if err != nil {
		return err
	}
	families, err := compileMarkFamilies(exp.descriptors)
	if err != nil {
		return err
	}
	if err := validateLaneAmbiguity(exp.descriptors); err != nil {
		return err
	}

	c.Root = root
	c.Codex = exp.codex
	c.families = families
	return nil
}

type explorer struct {
	tag         string // The root legend owns the namespace for every descriptor.
	codex       Codex
	journal     Journal
	descriptors []*markerDescriptor // Discovery order keeps finalization deterministic.
}

func (e *explorer) record(descriptor *markerDescriptor) {
	e.codex.Markers[descriptor.typ] = descriptor
	e.descriptors = append(e.descriptors, descriptor)
	e.journal.Visit(descriptor.typ)
}

func (e *explorer) explore(t reflect.Type) (*markerDescriptor, error) {
	original := t
	t = utilities.ResolvePointerChain(t)
	if t == nil || t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v: marker declaration must resolve to a struct", original)
	}
	if existing, ok := e.codex.Markers[t]; ok {
		return existing, nil
	}

	role, err := recognizeDeclarationRole(t)
	if err != nil {
		return nil, err
	}

	if err := validateMarkProvider(t, role); err != nil {
		return nil, err
	}

	descriptor, err := newMarkerDescriptor(t, role)
	if err != nil {
		return nil, err
	}
	lanes, err := e.inspectDeclaration(descriptor)
	if err != nil {
		return nil, err
	}

	surrogate := reflect.New(t).Interface()
	if err := descriptor.collectMark(e.tag, surrogate); err != nil {
		return nil, err
	}
	if err := descriptor.collectAcceptance(surrogate); err != nil {
		return nil, err
	}
	if role == legendRole {
		e.tag = descriptor.mark.tag
	}

	// Record locally validated descriptors before descending, so recursive spots
	// link to the same object. Chart.Explore publishes only the completed graph.
	e.record(descriptor)
	descriptor.lanes, err = e.collectLanes(t, lanes)
	if err != nil {
		return nil, err
	}
	if err := e.journal.Finish(t); err != nil {
		return nil, fmt.Errorf("%v: %w", t, err)
	}
	return descriptor, nil
}

// inspectDeclaration handles local validation and collection in declaration
// order. It prepares lanes without following their spots into other descriptors.
func (e *explorer) inspectDeclaration(descriptor *markerDescriptor) ([]laneField, error) {
	t := descriptor.typ
	features, qualifiers := make(map[string]bool), make(map[string]bool)
	var lanes []laneField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, tagged, err := readSailTag(field.Tag)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		if !tagged {
			continue
		}
		role := tag.Parts[0]
		if !field.IsExported() {
			return nil, fmt.Errorf("%v.%s: legend field must be exported", t, field.Name)
		}
		switch role {
		case "feature":
			if err := descriptor.collectFeature(field, features); err != nil {
				return nil, err
			}
		case "qualifier":
			if err := descriptor.collectQualifier(field, qualifiers); err != nil {
				return nil, err
			}
		case "barrel":
			if err := descriptor.collectBarrel(field); err != nil {
				return nil, err
			}
		case "lane":
			if descriptor.role == landmarkRole {
				return nil, fmt.Errorf("%v: a landmark cannot declare child lanes", t)
			}
			lane, err := inspectLane(field)
			if err != nil {
				return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
			}
			lanes = append(lanes, lane)
		default:
			return nil, fmt.Errorf("%v.%s: unexpected declaration role %q", t, field.Name, role)
		}
	}
	return lanes, nil
}

func (e *explorer) collectLanes(t reflect.Type, laneFields []laneField) ([]laneDescriptor, error) {
	var result []laneDescriptor
	for _, laneField := range laneFields {
		spots, err := e.collectSpots(laneField.typ)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, laneField.name, err)
		}
		result = append(result, laneDescriptor{
			declarationPath: laneField.declarationPath,
			strategy:        laneField.strategy,
			spots:           spots,
		})
	}
	return result, nil
}

func (e *explorer) collectSpots(t reflect.Type) ([]spotDescriptor, error) {
	// Local inspection has already resolved and validated this lane type.
	var result []spotDescriptor
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, tagged, err := readSailTag(field.Tag)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		if !tagged {
			continue
		}
		role := tag.Parts[0]
		if !field.IsExported() || role != "spot" || !isSpotField(field) {
			return nil, fmt.Errorf("%v.%s: expected an exported SingleSpot, OptionalSpot, or MultiSpot declaration", t, field.Name)
		}
		spotType := utilities.ResolvePointerChain(field.Type)
		if spotType == nil || spotType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%v.%s: spot must resolve to a struct", t, field.Name)
		}
		surrogate := reflect.New(spotType).Interface().(spotDeclaration)
		shape, err := utilities.InvokeDeclaration(spotType, "spot", surrogate.spot)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		// Validate M before normalization: pointer markers are not declaration
		// policy. OptionalSpot owns the pointer used to represent absence.
		if shape.target == nil || shape.target.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%v.%s: spot target %v must be a non-pointer marker struct", t, field.Name, shape.target)
		}
		carrier, err := carrierPath(spotType, shape.carrier)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		// Look up Value on the sealed carrier, never on a composite wrapper.
		valueField, _ := shape.carrier.FieldByName("Value")
		cardinality, err := compileCardinality(shape.kind, field.Tag)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		if utilities.TypeImplements[legendDeclaration](shape.target) {
			return nil, fmt.Errorf("%v.%s: spot must target a region or landmark, not a legend", t, field.Name)
		}
		targetDescriptor, err := e.explore(shape.target)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		result = append(result, spotDescriptor{
			declarationPath: append([]int(nil), field.Index...),
			valuePath:       slices.Concat(carrier, valueField.Index),
			kind:            shape.kind,
			cardinality:     cardinality,
			target:          targetDescriptor,
		})
	}
	return result, nil
}

func recognizeDeclarationRole(t reflect.Type) (markerRole, error) {
	var roles []markerRole
	if utilities.TypeImplements[legendDeclaration](t) {
		roles = append(roles, legendRole)
	}
	if utilities.TypeImplements[regionDeclaration](t) {
		roles = append(roles, regionRole)
	}
	if utilities.TypeImplements[landmarkDeclaration](t) {
		roles = append(roles, landmarkRole)
	}
	if len(roles) != 1 {
		return "", fmt.Errorf("%v: expected exactly one legend, region, or landmark role", t)
	}
	return roles[0], nil
}

// Required hooks are checked before allocating a surrogate, independently of
// marker role recognition.
func validateMarkProvider(t reflect.Type, role markerRole) error {
	if role == legendRole {
		if !utilities.TypeImplements[MarkProvider](t) {
			return fmt.Errorf("%v: legend declaration must provide Mark() string", t)
		}
	} else if !utilities.TypeImplements[FamilyProvider](t) {
		return fmt.Errorf("%v: region or landmark declaration must provide Family() string", t)
	}
	return nil
}
