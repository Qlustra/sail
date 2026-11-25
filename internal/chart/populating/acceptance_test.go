package populating

import (
	"reflect"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/rules/typerule"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/internal/chart/definitions"
	"github.com/qlustra/sail/legend"
)

type plain struct {
	legend.Landmark `sail:"landmark" family:"option"`
}
type boolOption struct {
	legend.Landmark `sail:"landmark" family:"option"`
}

var calls int

func (*boolOption) Accept() typerule.RuleAPI { calls++; return typerule.Type().Is[bool]() }

type hook struct{}

func (hook) Accept() typerule.RuleAPI { return typerule.Type().Is[bool]() }

type promoted struct {
	legend.Landmark `sail:"landmark" family:"option"`
	hook
}
type overridden struct {
	legend.Landmark `sail:"landmark" family:"option"`
	hook
}

func (overridden) Accept() typerule.RuleAPI { return typerule.Type().Is[string]() }

type panics struct {
	legend.Landmark `sail:"landmark" family:"option"`
}

func (panics) Accept() typerule.RuleAPI { panic("hook") }

type nilRule struct {
	legend.Landmark `sail:"landmark" family:"option"`
}

func (nilRule) Accept() typerule.RuleAPI { return nil }

type typedNil struct {
	legend.Landmark `sail:"landmark" family:"option"`
}

func (typedNil) Accept() typerule.RuleAPI { var r *typerule.Rule; return r }

type malformed struct {
	legend.Landmark `sail:"landmark" family:"option"`
}

func (malformed) Accept() typerule.RuleAPI { return typerule.Type() }

type incidentalRegion struct {
	legend.Region `sail:"region" family:"region"`
}

func (incidentalRegion) Accept() typerule.RuleAPI { panic("Region hook must not be called") }

type boolAlias = bool
type definedBool bool

func compileLandmark(t *testing.T, typ reflect.Type, success bool) *declarations.Landmark {
	t.Helper()
	d, err := define.Struct[definitions.LandmarkDefinition, declarations.Landmark](nil)
	if err != nil {
		t.Fatal(err)
	}
	s, c := capture.TypeReflection(typ).AsStruct()
	if !c.OK() {
		t.Fatal(c.Why())
	}
	ctx, v := d.Compile(s, nil)
	if v.IsOkay() != success {
		t.Fatalf("compile %v: %s", typ, v.Trace())
	}
	if !success {
		return nil
	}
	l, c := ctx.Access[declarations.Landmark](define.DeclarationKey)
	if !c.OK() {
		t.Fatal(c.Why())
	}
	return l
}

func TestAcceptance(t *testing.T) {
	calls = 0
	l := compileLandmark(t, reflect.TypeFor[boolOption](), true)
	if calls != 1 {
		t.Fatalf("hook calls = %d", calls)
	}
	for _, tc := range []struct {
		typ reflect.Type
		ok  bool
	}{
		{reflect.TypeFor[bool](), true}, {reflect.TypeFor[boolAlias](), true},
		{reflect.TypeFor[definedBool](), false}, {reflect.TypeFor[*bool](), false},
		{reflect.TypeFor[[]bool](), false}, {reflect.TypeFor[string](), false},
		{reflect.TypeFor[bool](), true},
	} {
		c := AcceptLandmark(l, capture.TypeReflection(tc.typ))
		if c.OK() != tc.ok {
			t.Errorf("%v: %s", tc.typ, c.Why())
		}
	}
	if calls != 1 {
		t.Fatalf("evaluation invoked hook: %d", calls)
	}
	if c := AcceptLandmark(compileLandmark(t, reflect.TypeFor[plain](), true), capture.Type[string]()); !c.OK() {
		t.Fatal(c.Why())
	}
	p := compileLandmark(t, reflect.TypeFor[promoted](), true)
	if c := AcceptLandmark(p, capture.Type[bool]()); !c.OK() {
		t.Fatal(c.Why())
	}
	o := compileLandmark(t, reflect.TypeFor[overridden](), true)
	if c := AcceptLandmark(o, capture.Type[string]()); !c.OK() {
		t.Fatal(c.Why())
	}
	for _, typ := range []reflect.Type{reflect.TypeFor[panics](), reflect.TypeFor[nilRule](), reflect.TypeFor[typedNil]()} {
		compileLandmark(t, typ, false)
	}
	m := compileLandmark(t, reflect.TypeFor[malformed](), true)
	if c := AcceptLandmark(m, capture.Type[bool]()); c.Code != "landmark_acceptance_evaluation_failed" {
		t.Fatalf("malformed: %s", c.Why())
	}
}

func TestRegionIgnoresAccept(t *testing.T) {
	d, err := define.Struct[definitions.RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	s, c := capture.Type[incidentalRegion]().AsStruct()
	if !c.OK() {
		t.Fatal(c.Why())
	}
	_, v := d.Compile(s, nil)
	if !v.IsOkay() {
		t.Fatal(v.Trace())
	}
}

func TestSharedBodyBindsOnce(t *testing.T) {
	d, err := define.Struct[definitions.LandmarkDefinition, declarations.Landmark](nil)
	if err != nil {
		t.Fatal(err)
	}
	s, c := capture.Type[boolOption]().AsStruct()
	if !c.OK() {
		t.Fatal(c.Why())
	}
	calls = 0
	first, v := d.Compile(s, nil)
	if !v.IsOkay() {
		t.Fatal(v.Trace())
	}
	second, v := d.Compile(s, first)
	if !v.IsOkay() {
		t.Fatal(v.Trace())
	}
	if first != second || calls != 1 {
		t.Fatalf("shared context %v, hook calls %d", first == second, calls)
	}
}
