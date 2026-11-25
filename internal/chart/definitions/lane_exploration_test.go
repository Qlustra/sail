package definitions

import (
	"reflect"
	"strings"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/legend"
)

// These fixtures check per-use metadata binding, not whole-grammar validity.
// Overlapping defaults are rejected by the separate completed-graph pass.
type policyLegendValid struct {
	legend.Legend `sail:"legend" mark:"x"`
	Default       sharedExplorationBody `sail:"lane"`
	Eager         sharedExplorationBody `sail:"lane" exploration:"eager"`
	Explicit      sharedExplorationBody `sail:"lane" exploration:"unambiguous"`
}
type policyRegionValid struct {
	legend.Region `sail:"region" family:"x"`
	Default       sharedExplorationBody `sail:"lane"`
	Eager         sharedExplorationBody `sail:"lane" exploration:"eager"`
	Explicit      sharedExplorationBody `sail:"lane" exploration:"unambiguous"`
}
type policyLegendEmpty struct {
	legend.Legend `sail:"legend" mark:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:""`
}
type policyLegendUnknown struct {
	legend.Legend `sail:"legend" mark:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"fast"`
}
type policyLegendCase struct {
	legend.Legend `sail:"legend" mark:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"Eager"`
}
type policyLegendLeading struct {
	legend.Legend `sail:"legend" mark:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:" eager"`
}
type policyLegendTrailing struct {
	legend.Legend `sail:"legend" mark:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"eager "`
}
type policyLegendCSV struct {
	legend.Legend `sail:"legend" mark:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"eager,unambiguous"`
}
type policyRegionEmpty struct {
	legend.Region `sail:"region" family:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:""`
}
type policyRegionUnknown struct {
	legend.Region `sail:"region" family:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"fast"`
}
type policyRegionCase struct {
	legend.Region `sail:"region" family:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"Eager"`
}
type policyRegionLeading struct {
	legend.Region `sail:"region" family:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:" eager"`
}
type policyRegionTrailing struct {
	legend.Region `sail:"region" family:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"eager "`
}
type policyRegionCSV struct {
	legend.Region `sail:"region" family:"x"`
	Use           sharedExplorationBody `sail:"lane" exploration:"eager,unambiguous"`
}

func TestLaneExplorationPolicies(t *testing.T) {
	cases := []struct {
		name  string
		typ   [2]reflect.Type
		valid bool
	}{
		{"valid", [2]reflect.Type{reflect.TypeFor[policyLegendValid](), reflect.TypeFor[policyRegionValid]()}, true},
		{"empty", [2]reflect.Type{reflect.TypeFor[policyLegendEmpty](), reflect.TypeFor[policyRegionEmpty]()}, false},
		{"unknown", [2]reflect.Type{reflect.TypeFor[policyLegendUnknown](), reflect.TypeFor[policyRegionUnknown]()}, false},
		{"case", [2]reflect.Type{reflect.TypeFor[policyLegendCase](), reflect.TypeFor[policyRegionCase]()}, false},
		{"leading", [2]reflect.Type{reflect.TypeFor[policyLegendLeading](), reflect.TypeFor[policyRegionLeading]()}, false},
		{"trailing", [2]reflect.Type{reflect.TypeFor[policyLegendTrailing](), reflect.TypeFor[policyRegionTrailing]()}, false},
		{"compound", [2]reflect.Type{reflect.TypeFor[policyLegendCSV](), reflect.TypeFor[policyRegionCSV]()}, false},
	}
	for _, region := range []bool{false, true} {
		for _, tc := range cases {
			t.Run(map[bool]string{false: "Legend/", true: "Region/"}[region]+tc.name, func(t *testing.T) {
				index := 0
				if region {
					index = 1
				}
				typ := tc.typ[index]
				subject, check := capture.TypeReflection(typ).AsStruct()
				if !check.OK() {
					t.Fatal(check.Why())
				}
				var lanes []*declarations.LaneCarrier
				if region {
					d, err := define.Struct[RegionDefinition, declarations.Region](nil)
					if err != nil {
						t.Fatal(err)
					}
					ctx, v := d.Compile(subject, nil)
					if !tc.valid {
						if v.IsOkay() || !strings.Contains(v.Trace(), "Use") || !strings.Contains(v.Trace(), "exploration") {
							t.Fatal(v.Trace())
						}
						return
					}
					if !v.IsOkay() {
						t.Fatal(v.Trace())
					}
					body, c := ctx.Access[declarations.Region](define.DeclarationKey)
					if !c.OK() {
						t.Fatal(c.Why())
					}
					lanes = body.Lanes
				} else {
					d, err := define.Struct[LegendDefinition, declarations.Legend](nil)
					if err != nil {
						t.Fatal(err)
					}
					ctx, v := d.Compile(subject, nil)
					if !tc.valid {
						if v.IsOkay() || !strings.Contains(v.Trace(), "Use") || !strings.Contains(v.Trace(), "exploration") {
							t.Fatal(v.Trace())
						}
						return
					}
					if !v.IsOkay() {
						t.Fatal(v.Trace())
					}
					body, c := ctx.Access[declarations.Legend](define.DeclarationKey)
					if !c.OK() {
						t.Fatal(c.Why())
					}
					lanes = body.Lanes
				}
				if len(lanes) != 3 || lanes[0].Exploration != declarations.UnambiguousExploration || lanes[1].Exploration != declarations.EagerExploration || lanes[2].Exploration != declarations.UnambiguousExploration {
					t.Fatalf("unexpected lanes: %+v", lanes)
				}
			})
		}
	}
}

type sharedExplorationBody struct {
	legend.Lane
}
type twoExplorationUses struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         sharedExplorationBody `sail:"lane" exploration:"eager"`
	Second        sharedExplorationBody `sail:"lane" exploration:"unambiguous"`
}
type invalidSecondExplorationUse struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         sharedExplorationBody `sail:"lane" exploration:"eager"`
	Second        sharedExplorationBody `sail:"lane" exploration:""`
}
type invalidSecondRegionExplorationUse struct {
	legend.Region `sail:"region" family:"x"`
	First         sharedExplorationBody `sail:"lane" exploration:"eager"`
	Second        sharedExplorationBody `sail:"lane" exploration:""`
}

func TestLegendLaneUsesRetainIndependentExploration(t *testing.T) {
	definition, err := define.Struct[LegendDefinition, declarations.Legend](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[twoExplorationUses]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := definition.Compile(subject, nil)
	if !validation.IsOkay() {
		t.Fatal(validation.Trace())
	}
	legendBody, check := ctx.Access[declarations.Legend](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(legendBody.Lanes) != 2 {
		t.Fatalf("got %d lane uses", len(legendBody.Lanes))
	}
	first, second := legendBody.Lanes[0], legendBody.Lanes[1]
	if first == second || first.Property == second.Property {
		t.Fatal("field uses must retain distinct wrappers and properties")
	}
	if first.Property.FieldName() != "First" || second.Property.FieldName() != "Second" {
		t.Fatalf("authored order/properties lost: %q, %q", first.Property.FieldName(), second.Property.FieldName())
	}
	if first.Exploration != declarations.EagerExploration || second.Exploration != declarations.UnambiguousExploration {
		t.Fatalf("policies = %q, %q", first.Exploration, second.Exploration)
	}
	if first.Declaration != second.Declaration {
		t.Fatal("field uses did not retain shared Lane body identity")
	}
	second.Exploration = declarations.EagerExploration
	trace := legendBody.Trace()
	if !strings.Contains(trace, `lane carrier "First" exploration="eager"`) || !strings.Contains(trace, `lane carrier "Second" exploration="eager"`) || !strings.Contains(trace, "(shared)") {
		t.Fatalf("trace did not use stored typed policy after mutation: %s", trace)
	}
}

func TestInvalidSecondRegionLaneUseIsNotAttached(t *testing.T) {
	d, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[invalidSecondRegionExplorationUse]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := d.Compile(subject, nil)
	if validation.IsOkay() || !strings.Contains(validation.Trace(), "Second") || !strings.Contains(validation.Trace(), "exploration") {
		t.Fatal(validation.Trace())
	}
	body, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(body.Lanes) != 1 || body.Lanes[0].Property.FieldName() != "First" {
		t.Fatal(body.Trace())
	}
}

func TestInvalidSecondLaneUseIsNotAttached(t *testing.T) {
	definition, err := define.Struct[LegendDefinition, declarations.Legend](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[invalidSecondExplorationUse]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := definition.Compile(subject, nil)
	if validation.IsOkay() {
		t.Fatal("compilation succeeded for invalid second lane use")
	}
	if !strings.Contains(validation.Trace(), "Second") || !strings.Contains(validation.Trace(), "exploration") {
		t.Fatalf("diagnostic omitted failing field: %s", validation.Trace())
	}
	legendBody, check := ctx.Access[declarations.Legend](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if len(legendBody.Lanes) != 1 || legendBody.Lanes[0].Property.FieldName() != "First" {
		t.Fatalf("invalid second use was attached: %s", legendBody.Trace())
	}
}
