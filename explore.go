package sail

import (
	"fmt"
	"reflect"

	"github.com/Qlustra/sail/internals/utilities"
)

// ======================================== Compiled schema

type World[C chart] struct {
	Root  *location
	Codex Codex

	families markFamilies
}

type Codex struct {
	Locations map[reflect.Type]*location
}

type lane struct {
	schemaField []int
	strategy    LaneExplorationStrategy
	spots       []spot
}

type spot struct {
	schemaField []int
	cardinality cardinality
	target      *location
}

// ======================================== Exploration

// Explore compiles C without inspecting or changing the live subject.
func Explore[C chart, S any](subject *S) (*World[C], *S, error) {
	world := &World[C]{}
	err := world.Explore()
	return world, subject, err
}

// Explore publishes a complete schema only after all locations and mark
// families have been validated. Failed compilation leaves the world unchanged.
func (w *World[C]) Explore() error {
	t := utilities.ResolvePointerChain(reflect.TypeFor[C]())
	if t == nil || t.Kind() != reflect.Struct || !utilities.TypeImplements[chartDeclaration](t) {
		return fmt.Errorf("%v is not a chart declaration", reflect.TypeFor[C]())
	}

	exp := explorer{
		codex:   Codex{Locations: make(map[reflect.Type]*location)},
		journal: NewJournal(),
	}
	root, err := exp.explore(t)
	if err != nil {
		return err
	}
	families, err := compileMarkFamilies(exp.locations)
	if err != nil {
		return err
	}

	w.Root = root
	w.Codex = exp.codex
	w.families = families
	return nil
}

type explorer struct {
	tag       string // The root chart owns the namespace for every location.
	codex     Codex
	journal   Journal
	locations []*location // Discovery order keeps finalization deterministic.
}

func (c *explorer) record(loc *location) {
	c.codex.Locations[loc.typ] = loc
	c.locations = append(c.locations, loc)
	c.journal.Visit(loc.typ)
}

func (c *explorer) explore(t reflect.Type) (*location, error) {
	original := t
	t = utilities.ResolvePointerChain(t)
	if t == nil || t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v: marker declaration must resolve to a struct", original)
	}
	if existing, ok := c.codex.Locations[t]; ok {
		return existing, nil
	}

	role, err := recognizeDeclarationRole(t)
	if err != nil {
		return nil, err
	}

	if err := validateMarkProvider(t, role); err != nil {
		return nil, err
	}

	location := newLocation(t, role)
	lanes, err := c.inspectDeclaration(location)
	if err != nil {
		return nil, err
	}

	surrogate := reflect.New(t).Interface()
	if err := location.collectMark(c.tag, surrogate); err != nil {
		return nil, err
	}
	if err := location.collectBehavior(surrogate); err != nil {
		return nil, err
	}
	if role == chartRole {
		c.tag = location.mark.tag
	}

	// Record locally validated locations before descending, so recursive spots
	// link to the same object. World.Explore publishes only the completed graph.
	c.record(location)
	location.lanes, err = c.collectLanes(t, lanes)
	if err != nil {
		return nil, err
	}
	if err := c.journal.Finish(t); err != nil {
		return nil, fmt.Errorf("%v: %w", t, err)
	}
	return location, nil
}

// laneField retains the local inspection needed for later lane expansion.
type laneField struct {
	name        string
	typ         reflect.Type
	schemaField []int
	strategy    LaneExplorationStrategy
}

// inspectDeclaration handles local validation and collection in declaration
// order. It prepares lanes without following their spots into other locations.
func (c *explorer) inspectDeclaration(location *location) ([]laneField, error) {
	t := location.typ
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
			return nil, fmt.Errorf("%v.%s: schema field must be exported", t, field.Name)
		}
		switch role {
		case "feature":
			if err := location.collectFeature(field, features); err != nil {
				return nil, err
			}
		case "qualifier":
			if err := location.collectQualifier(field, qualifiers); err != nil {
				return nil, err
			}
		case "lane":
			if location.role == landmarkRole {
				return nil, fmt.Errorf("%v: a landmark cannot declare child lanes", t)
			}
			if !isLaneField(field) {
				return nil, fmt.Errorf("%v.%s: lane must embed sail.Lane", t, field.Name)
			}
			laneType := utilities.ResolvePointerChain(field.Type)
			if laneType == nil || laneType.Kind() != reflect.Struct {
				return nil, fmt.Errorf("%v.%s: lane must resolve to a struct", t, field.Name)
			}
			value, err := readSingleTag(field.Tag, "exploration", nil)
			if err != nil {
				return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
			}
			strategy := LaneExplorationStrategy(value)
			switch strategy {
			case ExploreEagerly, ExploreConservatively, ExploreFreely:
			default:
				return nil, fmt.Errorf("%v.%s: invalid or missing exploration strategy %q", t, field.Name, strategy)
			}
			lanes = append(lanes, laneField{
				name:        field.Name,
				typ:         laneType,
				schemaField: append([]int(nil), field.Index...),
				strategy:    strategy,
			})
		default:
			return nil, fmt.Errorf("%v.%s: unexpected schema role %q", t, field.Name, role)
		}
	}
	return lanes, nil
}

func (c *explorer) collectLanes(t reflect.Type, laneFields []laneField) ([]lane, error) {
	var result []lane
	for _, laneField := range laneFields {
		spots, err := c.collectSpots(laneField.typ)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, laneField.name, err)
		}
		result = append(result, lane{
			schemaField: laneField.schemaField,
			strategy:    laneField.strategy,
			spots:       spots,
		})
	}
	return result, nil
}

func (c *explorer) collectSpots(t reflect.Type) ([]spot, error) {
	// Local inspection has already resolved and validated this lane type.
	var result []spot
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
			return nil, fmt.Errorf("%v.%s: expected an exported sail.Spot declaration", t, field.Name)
		}
		value, err := readSingleTag(field.Tag, "cardinality", nil)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		cardinality, err := compileCardinality(value)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		spotType := utilities.ResolvePointerChain(field.Type)
		if spotType == nil || spotType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%v.%s: spot must resolve to a struct", t, field.Name)
		}
		surrogate := reflect.New(spotType).Interface().(spotDeclaration)
		targetType, err := utilities.InvokeDeclaration(spotType, "Target", surrogate.Target)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		if utilities.TypeImplements[chartDeclaration](utilities.ResolvePointerChain(targetType)) {
			return nil, fmt.Errorf("%v.%s: spot must target a region or landmark, not a chart", t, field.Name)
		}
		targetLocation, err := c.explore(targetType)
		if err != nil {
			return nil, fmt.Errorf("%v.%s: %w", t, field.Name, err)
		}
		result = append(result, spot{
			schemaField: append([]int(nil), field.Index...),
			cardinality: cardinality,
			target:      targetLocation,
		})
	}
	return result, nil
}

func recognizeDeclarationRole(t reflect.Type) (locationRole, error) {
	var roles []locationRole
	if utilities.TypeImplements[chartDeclaration](t) {
		roles = append(roles, chartRole)
	}
	if utilities.TypeImplements[regionDeclaration](t) {
		roles = append(roles, regionRole)
	}
	if utilities.TypeImplements[landmarkDeclaration](t) {
		roles = append(roles, landmarkRole)
	}
	if len(roles) != 1 {
		return "", fmt.Errorf("%v: expected exactly one chart, region, or landmark role", t)
	}
	return roles[0], nil
}

// Required hooks are checked before allocating a surrogate. Spot.Target may
// be overridden, so reflected targets need validation beyond generic constraints.
func validateMarkProvider(t reflect.Type, role locationRole) error {
	if role == chartRole {
		if !utilities.TypeImplements[MarkProvider](t) {
			return fmt.Errorf("%v: chart declaration must provide Mark() string", t)
		}
	} else if !utilities.TypeImplements[FamilyProvider](t) {
		return fmt.Errorf("%v: region or landmark declaration must provide Family() string", t)
	}
	return nil
}
