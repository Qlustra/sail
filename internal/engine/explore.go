package engine

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qlustra/sail/internal/capability"
	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

type spotDescriptor struct {
	declarationPath []int
	valuePath       []int // Path inside the spot declaration to its embedded base's Value.
	kind            decl.SpotKind
	cardinality     cardinality
	target          *markerDescriptor
}

// ======================================== Compilation

// Explore compiles a complete charter before publishing any descriptors.
func Explore(legendType reflect.Type) (*Charter, error) {
	t := reflectutils.ResolvePointerChain(legendType)
	if t == nil || t.Kind() != reflect.Struct || !reflectutils.TypeImplements[decl.LegendRole](t) {
		return nil, fmt.Errorf("%v is not a legend declaration", legendType)
	}

	exp := explorer{
		codex:   Codex{Markers: make(map[reflect.Type]*markerDescriptor)},
		journal: NewJournal(),
	}
	root, err := exp.explore(t)
	if err != nil {
		return nil, err
	}
	families, err := compileMarkFamilies(exp.descriptors)
	if err != nil {
		return nil, err
	}
	if err := validateLaneAmbiguity(exp.descriptors); err != nil {
		return nil, err
	}

	return &Charter{Root: root, Codex: exp.codex, families: families}, nil
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
	t = reflectutils.ResolvePointerChain(t)
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
	// link to the same object. Charter.Compile publishes only the completed graph.
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
		if !field.IsExported() || role != "spot" || !reflectutils.FieldImplements[decl.SpotDeclaration](field) {
			return nil, fmt.Errorf("%v.%s: expected an exported SingleSpot, OptionalSpot, or MultiSpot declaration", t, field.Name)
		}
		spotType := reflectutils.ResolvePointerChain(field.Type)
		if spotType == nil || spotType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%v.%s: spot must resolve to a struct", t, field.Name)
		}
		if err := validateSpotEmbeddings(spotType); err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		surrogate := reflect.New(spotType).Interface().(decl.SpotDeclaration)
		shape, err := capability.InvokeDeclaration(spotType, "spot", func() decl.SpotShape { return decl.DescribeSpot(surrogate) })
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		// Validate M before normalization: pointer markers are not declaration
		// policy. OptionalSpot owns the pointer used to represent absence.
		if shape.Target == nil || shape.Target.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%v.%s: spot target %v must be a non-pointer marker struct", t, field.Name, shape.Target)
		}
		basePath, err := reflectutils.EmbeddedPath(spotType, shape.Base)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		// Look up Value on the sealed spot base, never on a composite wrapper.
		valueField, _ := shape.Base.FieldByName("Value")
		cardinality, err := compileCardinality(shape.Kind, field.Tag)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		if reflectutils.TypeImplements[decl.LegendRole](shape.Target) {
			return nil, fmt.Errorf("%v.%s: spot must target a region or landmark, not a legend", t, field.Name)
		}
		targetDescriptor, err := e.explore(shape.Target)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		result = append(result, spotDescriptor{
			declarationPath: append([]int(nil), field.Index...),
			valuePath:       slices.Concat(basePath, valueField.Index),
			kind:            shape.Kind,
			cardinality:     cardinality,
			target:          targetDescriptor,
		})
	}
	return result, nil
}

func recognizeDeclarationRole(t reflect.Type) (markerRole, error) {
	var roles []markerRole
	if reflectutils.TypeImplements[decl.LegendRole](t) {
		roles = append(roles, legendRole)
	}
	if reflectutils.TypeImplements[decl.RegionRole](t) {
		roles = append(roles, regionRole)
	}
	if reflectutils.TypeImplements[decl.LandmarkRole](t) {
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
		if !reflectutils.TypeImplements[decl.MarkProvider](t) {
			return fmt.Errorf("%v: legend declaration must provide Mark() string", t)
		}
	} else if !reflectutils.TypeImplements[decl.FamilyProvider](t) {
		return fmt.Errorf("%v: region or landmark declaration must provide Family() string", t)
	}
	return nil
}
