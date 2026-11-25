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

type discRegion struct {
	legend.Region `sail:"region" family:"f" discriminate:"z, a,A"`
}

func (discRegion) Discriminate() []string { return []string{"ignored"} }

type discLandmark struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:"z, a,A"`
}

func (discLandmark) Discriminate() []string { return []string{"ignored"} }

type emptyRegion struct {
	legend.Region `sail:"region" family:"f" discriminate:""`
}
type emptyLandmark struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:""`
}

func (emptyRegion) Discriminate() []string   { return []string{"ignored"} }
func (emptyLandmark) Discriminate() []string { return []string{"ignored"} }

type badRegion1 struct {
	legend.Region `sail:"region" family:"f" discriminate:"a,"`
}
type badRegion2 struct {
	legend.Region `sail:"region" family:"f" discriminate:",a"`
}
type badRegion3 struct {
	legend.Region `sail:"region" family:"f" discriminate:"a,,b"`
}
type badRegion4 struct {
	legend.Region `sail:"region" family:"f" discriminate:"a,a"`
}
type badLandmark1 struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:"a,"`
}
type badLandmark2 struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:",a"`
}
type badLandmark3 struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:"a,,b"`
}
type badLandmark4 struct {
	legend.Landmark `sail:"landmark" family:"f" discriminate:"a,a"`
}

func TestDiscriminatorBodies(t *testing.T) {
	r, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	l, err := define.Struct[LandmarkDefinition, declarations.Landmark](nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		typ           reflect.Type
		region, valid bool
		want          []string
	}{
		{reflect.TypeFor[discRegion](), true, true, []string{" a", "A", "z"}},
		{reflect.TypeFor[discLandmark](), false, true, []string{" a", "A", "z"}},
		{reflect.TypeFor[regionValid](), true, true, nil}, {reflect.TypeFor[landmarkValid](), false, true, nil},
		{reflect.TypeFor[emptyRegion](), true, true, nil}, {reflect.TypeFor[emptyLandmark](), false, true, nil},
		{reflect.TypeFor[badRegion1](), true, false, nil}, {reflect.TypeFor[badRegion2](), true, false, nil},
		{reflect.TypeFor[badRegion3](), true, false, nil}, {reflect.TypeFor[badRegion4](), true, false, nil},
		{reflect.TypeFor[badLandmark1](), false, false, nil}, {reflect.TypeFor[badLandmark2](), false, false, nil},
		{reflect.TypeFor[badLandmark3](), false, false, nil}, {reflect.TypeFor[badLandmark4](), false, false, nil},
	} {
		t.Run(tc.typ.Name(), func(t *testing.T) {
			subject, check := capture.TypeReflection(tc.typ).AsStruct()
			if !check.OK() {
				t.Fatal(check.Why())
			}
			var got []string
			var trace string
			if tc.region {
				ctx, v := r.Compile(subject, nil)
				if v.IsOkay() != tc.valid {
					t.Fatal(v.Reason())
				}
				if !tc.valid {
					return
				}
				body, check := ctx.Access[declarations.Region](define.DeclarationKey)
				if !check.OK() {
					t.Fatal(check.Why())
				}
				got, trace = body.Discriminators, body.Trace()
			} else {
				ctx, v := l.Compile(subject, nil)
				if v.IsOkay() != tc.valid {
					t.Fatal(v.Reason())
				}
				if !tc.valid {
					return
				}
				body, check := ctx.Access[declarations.Landmark](define.DeclarationKey)
				if !check.OK() {
					t.Fatal(check.Why())
				}
				got, trace = body.Discriminators, body.Trace()
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if len(got) > 0 && !strings.Contains(trace, `discriminate=" a,A,z"`) {
				t.Fatal(trace)
			}
		})
	}
}

type enumFlag struct {
	legend.Landmark `sail:"landmark" family:"flag" discriminate:"enum"`
}
type enumOther struct {
	legend.Landmark `sail:"landmark" family:"flag" discriminate:"other,enum"`
}
type orderedOne struct {
	legend.Region `sail:"region" family:"order" discriminate:"b,a"`
}
type orderedTwo struct {
	legend.Landmark `sail:"landmark" family:"order" discriminate:"a,b"`
}
type orderCollision struct {
	legend.Lane
	First  orderedOne   `sail:"spot" spot:"single"`
	Second []orderedTwo `sail:"spot" spot:"multi"`
}
type sameFlagLane struct {
	legend.Lane
	First  landmarkValid   `sail:"spot" spot:"single"`
	Second []landmarkValid `sail:"spot" spot:"multi"`
}
type differentLane struct {
	legend.Lane
	Plain landmarkValid `sail:"spot" spot:"single"`
	Enum  []enumFlag    `sail:"spot" spot:"multi"`
	Other enumOther     `sail:"spot" spot:"single"`
}
type flagLane2 struct {
	legend.Lane
	Flag landmarkValid `sail:"spot" spot:"single"`
}
type twoLanes struct {
	legend.Legend `sail:"legend" mark:"x"`
	One           differentLane `sail:"lane"`
	Two           flagLane2     `sail:"lane"`
}
type recursiveRegion struct {
	legend.Region `sail:"region" family:"r" discriminate:"enum"`
	Children      recursiveLane `sail:"lane"`
}
type recursiveLane struct {
	legend.Lane
	Back  []recursiveRegion `sail:"spot" spot:"multi"`
	Other recursiveOther    `sail:"spot" spot:"single"`
}
type recursiveOther struct {
	legend.Landmark `sail:"landmark" family:"r" discriminate:"enum"`
}
type recursiveValid struct {
	legend.Region `sail:"region" family:"r" discriminate:"enum"`
	Children      recursiveValidLane `sail:"lane"`
}
type recursiveValidLane struct {
	legend.Lane
	Back  []recursiveValid  `sail:"spot" spot:"multi"`
	Other recursiveDistinct `sail:"spot" spot:"single"`
}
type recursiveDistinct struct {
	legend.Landmark `sail:"landmark" family:"r" discriminate:"enum,other"`
}

func TestSelectorLaneCompilation(t *testing.T) {
	lane, err := define.Struct[LaneDefinition, declarations.Lane](nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		typ reflect.Type
		ok  bool
	}{{reflect.TypeFor[sameFlagLane](), false}, {reflect.TypeFor[orderCollision](), false}, {reflect.TypeFor[differentLane](), true}} {
		s, check := capture.TypeReflection(tc.typ).AsStruct()
		if !check.OK() {
			t.Fatal(check.Why())
		}
		ctx, v := lane.Compile(s, nil)
		if v.IsOkay() != tc.ok {
			t.Fatalf("%s: %s", tc.typ, v.Reason())
		}
		if !tc.ok && (!strings.Contains(v.Trace(), "duplicate_spot_selector") || !strings.Contains(v.Trace(), `fields "First" and "Second"`)) {
			t.Fatal(v.Trace())
		}
		// A failed session is not a usable product; inspect it only to verify attachment.
		body, check := ctx.Access[declarations.Lane](define.DeclarationKey)
		if !check.OK() {
			t.Fatal(check.Why())
		}
		trace := body.Trace()
		if !tc.ok {
			if strings.Count(trace, "spot ") != 1 || !strings.Contains(trace, "field First") || strings.Contains(trace, "field Second") {
				t.Fatal(trace)
			}
		} else {
			plain, enum, other := strings.Index(trace, "field Plain"), strings.Index(trace, "field Enum"), strings.Index(trace, "field Other")
			if plain < 0 || enum <= plain || other <= enum {
				t.Fatal(trace)
			}
		}
	}
	root, err := define.Struct[LegendDefinition, declarations.Legend](nil)
	if err != nil {
		t.Fatal(err)
	}
	s, check := capture.Type[twoLanes]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	_, v := root.Compile(s, nil)
	if !v.IsOkay() {
		t.Fatal(v.Reason())
	}
	r, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	rs, check := capture.Type[recursiveRegion]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	_, rv := r.Compile(rs, nil)
	if rv.IsOkay() {
		t.Fatal(rv.Trace())
	}
	// The enclosing Region currently summarizes nested compilation failures. The
	// Lane-rooted compilation exposes the checked collision and its field names.
	recursiveSubject, check := capture.Type[recursiveLane]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	_, collision := lane.Compile(recursiveSubject, nil)
	if collision.IsOkay() || !strings.Contains(collision.Trace(), "duplicate_spot_selector") || !strings.Contains(collision.Trace(), `fields "Back" and "Other"`) {
		t.Fatal(collision.Trace())
	}
	valid, check := capture.Type[recursiveValid]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	ctx, validation := r.Compile(valid, nil)
	if !validation.IsOkay() {
		t.Fatal(validation.Trace())
	}
	body, check := ctx.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if trace := body.Trace(); !strings.Contains(trace, `family="r" discriminate="enum"`) || !strings.Contains(trace, "(recursive)") {
		t.Fatal(trace)
	}
}
