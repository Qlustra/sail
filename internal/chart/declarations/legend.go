package declarations

import (
	"fmt"
	"strings"

	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/inspect/tag/formats"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules/typerule"
	"github.com/qlustra/sail/internal/chart/carriers"
)

// ================================================== Legend

// The completed-graph namespace check covers reachable bodies per family.
type Legend struct {
	define.ScopeDeclaration
	Mark       string
	Lanes      []*LaneCarrier
	Warnings   []Warning
	Vocabulary *Vocabulary
}

func (l *Legend) AddLane(lane *LaneCarrier) logic.Check {
	return addLane(&l.Lanes, lane)
}

func (l *Legend) String() string {
	return l.Mark
}

func (l *Legend) Trace() string {
	return traceDeclaration(l)
}

// ================================================== Lane

type Lane struct {
	define.ScopeDeclaration
	Mark string
	// Discriminator/Qualifier conflicts are checked on the completed graph, not per Lane.
	spots     []*Spot
	selectors map[selector]*Spot
}

type selector struct{ family, discriminators string }

// Spots exposes the compiled members in authored order. Population keeps its
// occurrence counts separately; it must not mutate this shared list or its bodies.
func (l *Lane) Spots() []*Spot { return l.spots }

// Select looks up the exact family and discriminator set; qualifiers do not select.
func (l *Lane) Select(mark *ChartMark) *Spot {
	if l == nil || mark == nil {
		return nil
	}
	return l.selectors[mark.selector]
}

func spotSelector(spot *Spot) selector {
	var family string
	var discriminators []string
	if spot.For == SpotForRegion {
		family, discriminators = spot.Region.Declaration.Family, spot.Region.Declaration.Discriminators
	} else {
		family, discriminators = spot.Landmark.Declaration.Family, spot.Landmark.Declaration.Discriminators
	}
	// CSV components cannot contain commas, so joining normalized components is
	// unambiguous; the family is kept in a separate key field.
	return selector{family, strings.Join(discriminators, ",")}
}

func (l *Lane) AddSpot(spot *Spot) logic.Check {
	key := spotSelector(spot)
	if previous := l.selectors[key]; previous != nil {
		return logic.Nope("duplicate_spot_selector", fmt.Sprintf("fields %q and %q have duplicate selector family=%q discriminate=%q", spotField(previous), spotField(spot), key.family, key.discriminators))
	}
	if l.selectors == nil {
		l.selectors = make(map[selector]*Spot)
	}
	l.selectors[key] = spot
	l.spots = append(l.spots, spot)
	return logic.Ok()
}

func spotField(s *Spot) string {
	if s.Region != nil && s.Region.Property != nil {
		return s.Region.Property.FieldName()
	}
	if s.Landmark != nil && s.Landmark.Property != nil {
		return s.Landmark.Property.FieldName()
	}
	return "<unknown>"
}

func (l *Lane) String() string {
	return l.Mark
}

func (l *Lane) Trace() string {
	return traceDeclaration(l)
}

// ================================================== Spot

type SpotKind string
type SpotFor string

const (
	SingleSpot SpotKind = "single"
	MultiSpot  SpotKind = "multi"
)

const (
	SpotForRegion   SpotFor = "region"
	SpotForLandmark SpotFor = "landmark"
)

// Spot is a vetted lane-member view over a Region or Landmark carrier, not a
// separately compiled body declaration.
type Spot struct {
	Required bool
	Kind     SpotKind
	For      SpotFor
	Region   *define.Carrier[Region]
	Landmark *define.Carrier[Landmark]
}

func (s *Spot) String() string {
	return fmt.Sprintf("%s %s", s.Kind, s.For)
}

func (s *Spot) Trace() string {
	return traceDeclaration(s)
}

type SpotTag struct {
	formats.CSV
}

func (t *SpotTag) Get() SpotKind {
	return SpotKind(t.Parts[0])
}

func (t SpotTag) Validate() logic.Check {
	// Representation is mandatory and first; only one required qualifier is valid.
	if len(t.Parts) == 0 || len(t.Parts) > 2 {
		return logic.Nope("invalid spot tag", t.Raw)
	}
	kind := SpotKind(t.Parts[0])
	if kind != SingleSpot && kind != MultiSpot {
		return logic.Nope("invalid spot tag", t.Raw)
	}
	if len(t.Parts) == 2 && t.Parts[1] != "required" {
		return logic.Nope("invalid spot tag", t.Raw)
	}
	return logic.Ok()
}

func (t *SpotTag) Required() bool { return len(t.Parts) == 2 }

// ================================================== Region

type Region struct {
	define.ScopeDeclaration
	Family         string
	Discriminators []string
	Features       []*carriers.Feature
	Barrels        []*carriers.Barrel
	Qualifiers     []*QualifierCarrier
	Lanes          []*LaneCarrier
}

func (l *Region) AddFeatures(feature *carriers.Feature) logic.Check {
	var check logic.Check
	l.Features, check = addFeature(l.Features, feature)
	return check
}

func (l *Region) AddQualifier(qualifier *QualifierCarrier) logic.Check {
	var check logic.Check
	l.Qualifiers, check = addQualifier(l.Qualifiers, qualifier)
	return check
}

func (l *Region) AddLane(lane *LaneCarrier) logic.Check {
	return addLane(&l.Lanes, lane)
}

// ================================================== Landmark

type Landmark struct {
	define.ScopeDeclaration
	Family         string
	Discriminators []string
	Features       []*carriers.Feature
	Barrels        []*carriers.Barrel
	Qualifiers     []*QualifierCarrier
	Acceptance     typerule.RuleAPI
}

func (l *Landmark) AddQualifier(qualifier *QualifierCarrier) logic.Check {
	var check logic.Check
	l.Qualifiers, check = addQualifier(l.Qualifiers, qualifier)
	return check
}

func (l *Landmark) AddFeatures(feature *carriers.Feature) logic.Check {
	var check logic.Check
	l.Features, check = addFeature(l.Features, feature)
	return check
}

// ================================================== Qualifier

type Qualifier struct {
	define.ScopeDeclaration
	Barrels []*carriers.Barrel
}
