package definitions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/inspect"
	"github.com/qlustra/intel/inspect/tag/formats"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/options"
	"github.com/qlustra/intel/rules/extension"
	"github.com/qlustra/intel/rules/field"
	"github.com/qlustra/intel/rules/property"
	"github.com/qlustra/intel/rules/typerule"
	"github.com/qlustra/sail/internal/chart/behavior"
	"github.com/qlustra/sail/internal/chart/carriers"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/values"
	"github.com/qlustra/sail/legend"
)

// ================================================== Legend

type LegendDefinition struct {
	MarkerDefinition[declarations.Legend]
}

func (d *LegendDefinition) Define() {
	d.IgnoreDescendantsOf[legend.Legend]()
	d.Marker[behavior.Legend](
		"legend",
		extension.Field(field.TaggedAsSingle("mark")),
	).Encounter(func(encounter extension.Encounter) logic.Check {
		legendDeclaration, check := encounter.Access[declarations.Legend](define.DeclarationKey)
		if !check.OK() {
			return check
		}
		legendDeclaration.Mark = encounter.Extension.Tags().Raw("mark")
		return logic.Ok()
	})

	d.Scope[declarations.Lane, LaneDefinition](
		property.Satisfies[behavior.Lane](),
		property.Field(field.TaggedAsSingleValue("sail", "lane")),
	).Encounter(func(
		encounter property.Encounter,
		legend *declarations.Legend,
		lane *define.Carrier[declarations.Lane],
	) logic.Check {
		use, check := declarations.NewLaneCarrier(lane)
		if !check.OK() {
			return check
		}
		return legend.AddLane(use)
	})
}

// ================================================== Lane

type LaneDefinition struct {
	MarkerDefinition[declarations.Lane]
}

func (d *LaneDefinition) Define() {
	d.IgnoreDescendantsOf[legend.Lane]()
	d.Scope[declarations.Region, RegionDefinition](
		property.Satisfies[behavior.Region](),
		property.Field(field.TaggedAsSingleValue("sail", "spot")),
	).Target(spotTarget).Encounter(func(
		encounter property.Encounter,
		lane *declarations.Lane,
		region *define.Carrier[declarations.Region],
	) logic.Check {
		spot, check := regionSpot(region, declarations.SingleSpot)
		if !check.OK() {
			return check
		}
		return lane.AddSpot(spot)
	})

	d.Scope[declarations.Region, RegionDefinition](
		property.StructSlice(typerule.Type().Satisfies[behavior.Region]()),
		property.Field(field.TaggedAsSingleValue("sail", "spot")),
	).Target(spotSliceTarget).Encounter(func(
		encounter property.Encounter,
		lane *declarations.Lane,
		region *define.Carrier[declarations.Region],
	) logic.Check {
		spot, check := regionSpot(region, declarations.MultiSpot)
		if !check.OK() {
			return check
		}
		return lane.AddSpot(spot)
	})

	d.Leaf[declarations.Landmark, LandmarkDefinition](
		property.Satisfies[behavior.Landmark](),
		property.Field(field.TaggedAsSingleValue("sail", "spot")),
	).Encounter(func(
		encounter property.Encounter,
		lane *declarations.Lane,
		landmark *define.Carrier[declarations.Landmark],
	) logic.Check {
		spot, check := landmarkSpot(landmark, declarations.SingleSpot)
		if !check.OK() {
			return check
		}
		return lane.AddSpot(spot)
	})

	d.Leaf[declarations.Landmark, LandmarkDefinition](
		property.StructSlice(typerule.Type().Satisfies[behavior.Landmark]()),
		property.Field(field.TaggedAsSingleValue("sail", "spot")),
	).Target(spotSliceTarget).Encounter(func(
		encounter property.Encounter,
		lane *declarations.Lane,
		landmark *define.Carrier[declarations.Landmark],
	) logic.Check {
		spot, check := landmarkSpot(landmark, declarations.MultiSpot)
		if !check.OK() {
			return check
		}
		return lane.AddSpot(spot)
	})
}

// ================================================== Region

type RegionDefinition struct {
	MarkerDefinition[declarations.Region]
}

func (d *RegionDefinition) Define() {
	d.IgnoreDescendantsOf[legend.Region]()
	// Defined marker types copy their embedded cell without a legend.Region
	// ancestor. That exact internal state type is also library-owned.
	d.IgnoreDescendantsOf[values.Occurrence]()
	d.Marker[behavior.Region]("region", familyTag()).Encounter(func(encounter extension.Encounter) logic.Check {
		region, check := encounter.Access[declarations.Region](define.DeclarationKey)
		if !check.OK() {
			return check
		}
		region.Family = encounter.Extension.Tags().Raw("family")
		region.Discriminators, check = discriminatorTag(encounter.Extension.Tags())
		if !check.OK() {
			return check
		}
		return logic.Ok()
	})

	d.Carrier[behavior.Feature](property.Field(field.TaggedAsSingleValue("sail", "feature"))).Encounter(func(
		encounter property.Encounter,
		region *declarations.Region,
	) logic.Check {
		feature, check := carriers.NewFeature(&encounter)
		if !check.OK() {
			return check
		}
		return region.AddFeatures(feature)
	})

	d.Carrier[behavior.Barrel](property.Field(field.TaggedAsSingleValue("sail", "barrel"))).Encounter(func(
		encounter property.Encounter,
		region *declarations.Region,
	) logic.Check {
		barrel, check := carriers.NewBarrel(&encounter)
		if !check.OK() {
			return check
		}
		region.Barrels = append(region.Barrels, barrel)
		return logic.Ok()
	})

	d.Property(options.Optional, bareQualifier(), property.Field(field.TaggedAsSingleValue("sail", "qualifier"))).Encounter(func(e property.Encounter) logic.Check {
		host, check := e.Access[declarations.Region](define.DeclarationKey)
		if !check.OK() {
			return check
		}
		use, check := declarations.NewQualifierCarrier(&e, nil)
		if !check.OK() {
			return check
		}
		return host.AddQualifier(use)
	})
	d.Leaf[declarations.Qualifier, QualifierDefinition](
		property.And(property.Satisfies[behavior.Qualifier](), property.Not(bareQualifier())),
		property.Field(field.TaggedAsSingleValue("sail", "qualifier")),
	).Encounter(func(e property.Encounter, host *declarations.Region, body *define.Carrier[declarations.Qualifier]) logic.Check {
		return attachQualifier(e, body, host.AddQualifier)
	})

	d.Scope[declarations.Lane, LaneDefinition](
		property.Satisfies[behavior.Lane](),
		property.Field(field.TaggedAsSingleValue("sail", "lane")),
	).Encounter(func(
		encounter property.Encounter,
		region *declarations.Region,
		lane *define.Carrier[declarations.Lane],
	) logic.Check {
		use, check := declarations.NewLaneCarrier(lane)
		if !check.OK() {
			return check
		}
		return region.AddLane(use)
	})
}

// ================================================== Landmark

type LandmarkDefinition struct {
	MarkerDefinition[declarations.Landmark]
}

func (d *LandmarkDefinition) Define() {
	d.IgnoreDescendantsOf[legend.Landmark]()
	d.IgnoreDescendantsOf[values.Occurrence]()
	d.Marker[behavior.Landmark]("landmark", familyTag()).Encounter(func(encounter extension.Encounter) logic.Check {
		landmark, check := encounter.Access[declarations.Landmark](define.DeclarationKey)
		if !check.OK() {
			return check
		}
		landmark.Family = encounter.Extension.Tags().Raw("family")
		landmark.Discriminators, check = discriminatorTag(encounter.Extension.Tags())
		if !check.OK() {
			return check
		}
		// Bind on the outer authored body, before properties can link recursively.
		rule, check := encounter.Struct.CreateValue().Invoke("Accept", func(h interface{ Accept() typerule.RuleAPI }) (typerule.RuleAPI, logic.Check) {
			return h.Accept(), logic.Ok()
		})
		if check.Code == capture.ValueTypeMismatchCode {
			return logic.Ok()
		}
		if !check.OK() {
			return check
		}
		if inspect.IsNil(rule) {
			return logic.Nope("nil_landmark_acceptance", "Accept returned a nil rule")
		}
		landmark.Acceptance = rule
		return logic.Ok()
	})

	d.Carrier[behavior.Feature](property.Field(field.TaggedAsSingleValue("sail", "feature"))).Encounter(func(
		encounter property.Encounter,
		landmark *declarations.Landmark,
	) logic.Check {
		feature, check := carriers.NewFeature(&encounter)
		if !check.OK() {
			return check
		}
		return landmark.AddFeatures(feature)
	})

	d.Carrier[behavior.Barrel](property.Field(field.TaggedAsSingleValue("sail", "barrel"))).Encounter(func(
		encounter property.Encounter,
		landmark *declarations.Landmark,
	) logic.Check {
		barrel, check := carriers.NewBarrel(&encounter)
		if !check.OK() {
			return check
		}
		landmark.Barrels = append(landmark.Barrels, barrel)
		return logic.Ok()
	})

	d.Property(options.Optional, bareQualifier(), property.Field(field.TaggedAsSingleValue("sail", "qualifier"))).Encounter(func(e property.Encounter) logic.Check {
		host, check := e.Access[declarations.Landmark](define.DeclarationKey)
		if !check.OK() {
			return check
		}
		use, check := declarations.NewQualifierCarrier(&e, nil)
		if !check.OK() {
			return check
		}
		return host.AddQualifier(use)
	})
	d.Leaf[declarations.Qualifier, QualifierDefinition](
		property.And(property.Satisfies[behavior.Qualifier](), property.Not(bareQualifier())),
		property.Field(field.TaggedAsSingleValue("sail", "qualifier")),
	).Encounter(func(e property.Encounter, host *declarations.Landmark, body *define.Carrier[declarations.Qualifier]) logic.Check {
		return attachQualifier(e, body, host.AddQualifier)
	})
}

// Family is one consumer CSV token: it must be nonempty and cannot contain the
// delimiter. No further identifier syntax is imposed; text is retained verbatim.
func familyTag() extension.RuleAPI {
	return extension.Field(field.TaggedAs[formats.Single]("family", func(tag *inspect.Tag[formats.Single]) bool {
		return tag.Format.Raw != "" && !strings.Contains(tag.Format.Raw, ",")
	}))
}

// Selector metadata must be finalized in the marker phase: nested properties can
// link back to this body while it is still being compiled. Discriminator
// components are only required to be nonempty and unique; their
// text is otherwise verbatim, with no additional lexical or case rules.
func discriminatorTag(tags *inspect.FieldTagInspector) ([]string, logic.Check) {
	if tags.Raw("discriminate") == "" {
		return nil, logic.Ok()
	}
	tag, check := tags.Parse[formats.CSV]("discriminate")
	if !check.OK() {
		return nil, check
	}
	parts := tag.Format.Parts
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		if part == "" || seen[part] {
			return nil, logic.Nope("invalid_discriminator", fmt.Sprintf("discriminate contains an empty or repeated component %q", part))
		}
		seen[part] = true
	}
	slices.Sort(parts)
	return parts, logic.Ok()
}

// ================================================== Qualifier

// A bare Qualifier is presence-only. A custom body has a matching marker and
// one or more Barrel fields. Its symbol belongs to the incoming field use.
type QualifierDefinition struct {
	MarkerDefinition[declarations.Qualifier]
}

func bareQualifier() *property.Rule {
	rule := property.Property()
	rule.ExpectCheck("bare Qualifier", func(e property.Encounter) logic.Check {
		// Type, not declaration depth, distinguishes the primitive. Promoted
		// qualifier fields remain bodyless just like directly declared ones.
		return e.Property.Is[legend.Qualifier]()
	})
	return rule
}

func (d *QualifierDefinition) Define() {
	d.IgnoreDescendantsOf[legend.Qualifier]()
	d.Marker[behavior.Qualifier]("qualifier").Encounter(func(e extension.Encounter) logic.Check {
		// Intel's closed property composition does not cover other embeddings or
		// blank fields. Inspect direct fields so promoted marker internals are not
		// mistaken for authored members.
		marker := false
		for name, field := range e.Struct.Fields.AllDeclared() {
			if field.IsEmbedding().OK() {
				if !field.Is[legend.Qualifier]().OK() {
					return logic.Nope("invalid_qualifier_member", fmt.Sprintf("field %s: only the Qualifier marker may be embedded", name))
				}
				marker = true
			} else if field.IsBlank().OK() {
				return logic.Nope("invalid_qualifier_member", "blank fields are not Barrel properties")
			}
		}
		if !marker {
			return logic.Nope("missing_qualifier_marker", "custom qualifier requires a directly embedded marker")
		}
		return logic.Ok()
	})
	d.Carrier[behavior.Barrel](property.Field(field.TaggedAsSingleValue("sail", "barrel"))).Encounter(func(e property.Encounter, body *declarations.Qualifier) logic.Check {
		barrel, check := carriers.NewBarrel(&e)
		if !check.OK() {
			return check
		}
		body.Barrels = append(body.Barrels, barrel)
		return logic.Ok()
	})
}

func attachQualifier(e property.Encounter, body *define.Carrier[declarations.Qualifier], add func(*declarations.QualifierCarrier) logic.Check) logic.Check {
	if len(body.Declaration.Barrels) == 0 {
		return logic.Nope("empty_qualifier_body", fmt.Sprintf("field %s: custom qualifier requires at least one Barrel", e.Property.FieldName()))
	}
	use, check := declarations.NewQualifierCarrier(&e, body.Declaration)
	if !check.OK() {
		return check
	}
	return add(use)
}
