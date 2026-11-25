package declarations

import (
	"fmt"
	"strings"

	"github.com/qlustra/sail/internal/chart/carriers"
)

// Each trace has its own discovery-order identities. A declaration is expanded
// once; later edges identify either a recursive link or a shared declaration.
type declarationTrace struct {
	output strings.Builder
	seen   map[any]int
	active map[any]bool
}

func traceDeclaration(root any) string {
	trace := declarationTrace{
		seen:   make(map[any]int),
		active: make(map[any]bool),
	}
	trace.walk(root, 0)
	return strings.TrimSuffix(trace.output.String(), "\n")
}

func (t *declarationTrace) line(depth int, label string) {
	t.output.WriteString(strings.Repeat("  ", depth))
	t.output.WriteString(label)
	t.output.WriteByte('\n')
}

func (t *declarationTrace) walk(declaration any, depth int) {
	label, present := declarationLabel(declaration)
	if !present {
		t.line(depth, label)
		return
	}
	if identity, seen := t.seen[declaration]; seen {
		relation := "shared"
		if t.active[declaration] {
			relation = "recursive"
		}
		t.line(depth, fmt.Sprintf("%s -> #%d (%s)", label, identity, relation))
		return
	}
	identity := len(t.seen) + 1
	t.seen[declaration] = identity
	t.active[declaration] = true
	t.line(depth, fmt.Sprintf("%s #%d", label, identity))

	switch value := declaration.(type) {
	case *Legend:
		for _, lane := range value.Lanes {
			t.walkLane(lane, depth+1)
		}
	case *Lane:
		for _, spot := range value.spots {
			t.walk(spot, depth+1)
		}
	case *Spot:
		switch value.For {
		case SpotForRegion:
			if value.Region != nil {
				t.walk(value.Region.Declaration, depth+1)
			} else {
				t.line(depth+1, "region carrier <nil>")
			}
		case SpotForLandmark:
			if value.Landmark != nil {
				t.walk(value.Landmark.Declaration, depth+1)
			} else {
				t.line(depth+1, "landmark carrier <nil>")
			}
		default:
			if value.Region != nil {
				t.walk(value.Region.Declaration, depth+1)
			}
			if value.Landmark != nil {
				t.walk(value.Landmark.Declaration, depth+1)
			}
		}
	case *Region:
		for _, feature := range value.Features {
			t.walk(feature, depth+1)
		}
		for _, barrel := range value.Barrels {
			t.walk(barrel, depth+1)
		}
		for _, qualifier := range value.Qualifiers {
			t.walkQualifier(qualifier, depth+1)
		}
		for _, lane := range value.Lanes {
			t.walkLane(lane, depth+1)
		}
	case *Landmark:
		for _, feature := range value.Features {
			t.walk(feature, depth+1)
		}
		for _, barrel := range value.Barrels {
			t.walk(barrel, depth+1)
		}
		for _, qualifier := range value.Qualifiers {
			t.walkQualifier(qualifier, depth+1)
		}
	case *Qualifier:
		for _, barrel := range value.Barrels {
			t.walk(barrel, depth+1)
		}
	}
	delete(t.active, declaration)
}

func (t *declarationTrace) walkQualifier(carrier *QualifierCarrier, depth int) {
	if carrier == nil {
		t.line(depth, "qualifier field <nil>")
		return
	}
	label := fmt.Sprintf("qualifier field %q", carrier.Symbol)
	if carrier.Property != nil {
		label += fmt.Sprintf(" (field %s)", carrier.Property.FieldName())
	}
	if carrier.Declaration == nil {
		t.line(depth, label+" presence-only")
	} else {
		t.line(depth, label+" custom")
		t.walk(carrier.Declaration, depth+1)
	}
}

func (t *declarationTrace) walkLane(carrier *LaneCarrier, depth int) {
	if carrier == nil {
		t.line(depth, "lane carrier <nil>")
		return
	}
	label := "lane carrier"
	if carrier.Carrier != nil && carrier.Property != nil {
		label += fmt.Sprintf(" %q", carrier.Property.FieldName())
	}
	label += fmt.Sprintf(" exploration=%q", carrier.Exploration)
	t.line(depth, label)
	if carrier.Carrier == nil {
		t.walk((*Lane)(nil), depth+1)
	} else {
		t.walk(carrier.Declaration, depth+1)
	}
}

func declarationLabel(declaration any) (string, bool) {
	switch value := declaration.(type) {
	case *Legend:
		if value == nil {
			return "legend <nil>", false
		}
		return fmt.Sprintf("legend %q", value.Mark), true
	case *Lane:
		if value == nil {
			return "lane <nil>", false
		}
		if value.Mark != "" {
			return fmt.Sprintf("lane %q", value.Mark), true
		}
		return "lane", true
	case *Spot:
		if value == nil {
			return "spot <nil>", false
		}
		label := "spot " + value.String()
		if value.Region != nil && value.Region.Property != nil {
			label += fmt.Sprintf(" (field %s)", value.Region.Property.FieldName())
		} else if value.Landmark != nil && value.Landmark.Property != nil {
			label += fmt.Sprintf(" (field %s)", value.Landmark.Property.FieldName())
		}
		if value.Required {
			label += " (required)"
		}
		return label, true
	case *Region:
		if value == nil {
			return "region <nil>", false
		}
		return bodyLabel("region", value.Family, value.Discriminators), true
	case *Landmark:
		if value == nil {
			return "landmark <nil>", false
		}
		label := bodyLabel("landmark", value.Family, value.Discriminators)
		if value.Acceptance != nil {
			label += " acceptance=present"
		}
		return label, true
	case *carriers.Feature:
		return value.String(), value != nil
	case *carriers.Barrel:
		if value == nil {
			return "barrel <nil>", false
		}
		label := fmt.Sprintf("barrel required=%t", value.Required)
		if value.Encounter != nil && value.Encounter.Property != nil {
			label += fmt.Sprintf(" (field %s)", value.Encounter.Property.FieldName())
		}
		label += fmt.Sprintf(" populate=%q capability=%v result=%v", value.Populate, value.Shape.Capability.CapturedType(), value.Shape.Result.CapturedType())
		return label, true
	case *Qualifier:
		return "qualifier" + nilSuffix(value == nil), value != nil
	default:
		return fmt.Sprintf("unknown declaration %T", declaration), false
	}
}

func bodyLabel(kind, family string, discriminators []string) string {
	label := fmt.Sprintf("%s family=%q", kind, family)
	if len(discriminators) != 0 {
		label += fmt.Sprintf(" discriminate=%q", strings.Join(discriminators, ","))
	}
	return label
}

func nilSuffix(missing bool) string {
	if missing {
		return " <nil>"
	}
	return ""
}

func (r *Region) Trace() string    { return traceDeclaration(r) }
func (l *Landmark) Trace() string  { return traceDeclaration(l) }
func (q *Qualifier) Trace() string { return traceDeclaration(q) }
