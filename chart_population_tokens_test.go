package sail_test

import (
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

var tokenCalls atomic.Int64

type tokenSelf int

func (tokenSelf) Default() tokenSelf { tokenCalls.Add(1); return 17 }

type tokenOther int

func (tokenOther) Default() string { tokenCalls.Add(1); return "other" }

type tokenPanic int

func (tokenPanic) Default() tokenPanic { tokenCalls.Add(1); panic("token getter panic") }

type tokenCustom struct {
	legend.Qualifier `sail:"qualifier"`
	Value            legend.Barrel[any, any] `sail:"barrel" populate:"Default" barrel:"required"`
	Optional         legend.Barrel[any, any] `sail:"barrel" populate:"Unavailable"`
}

type TokenPromotedQualifiers struct {
	Active legend.Qualifier `sail:"qualifier" symbol:"active"`
	First  tokenCustom      `sail:"qualifier" symbol:"first"`
}

type tokenPlain struct {
	legend.Landmark `sail:"landmark" family:"flag"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
	TokenPromotedQualifiers
	Second tokenCustom `sail:"qualifier" symbol:"second"`
}
type tokenEnum struct {
	legend.Landmark `sail:"landmark" family:"flag" discriminate:"enum"`
	Name            legend.Feature   `sail:"feature" symbol:"name"`
	EnumOnly        legend.Qualifier `sail:"qualifier" symbol:"enumonly"`
}
type tokenPair struct {
	legend.Landmark `sail:"landmark" family:"flag" discriminate:"enum,multi"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}
type tokenPosition struct {
	legend.Landmark `sail:"landmark" family:"active"`
	Active          legend.Qualifier `sail:"qualifier" symbol:"active"`
}
type tokenLane struct {
	legend.Lane
	Plain    []tokenPlain    `sail:"spot" spot:"multi"`
	Enum     []tokenEnum     `sail:"spot" spot:"multi"`
	Pair     []tokenPair     `sail:"spot" spot:"multi"`
	Position []tokenPosition `sail:"spot" spot:"multi"`
}
type tokenFallback struct {
	legend.Landmark `sail:"landmark" family:"flag"`
	LateOnly        legend.Qualifier `sail:"qualifier" symbol:"lateonly"`
}
type tokenFallbackLane struct {
	legend.Lane
	Flag tokenFallback `sail:"spot" spot:"single"`
}
type tokenRoot struct {
	legend.Legend `sail:"legend" mark:"tokens"`
	Items         tokenLane         `sail:"lane" exploration:"eager"`
	Fallback      tokenFallbackLane `sail:"lane"`
}

type TokenPromotedRecipe struct {
	P tokenSelf `tokens:"flag,first,active" name:"promoted"`
}
type tokenRecipe struct {
	TokenPromotedRecipe
	E        bool       `tokens:"flag,enum,enumonly" name:"enum"`
	A        tokenOther `tokens:"flag,second" name:"other"`
	B        tokenSelf  `tokens:"flag" name:"bare"`
	C        tokenSelf  `tokens:"flag,first,second" name:"both"`
	Pair     bool       `tokens:"flag,multi,enum" name:"pair"`
	Position bool       `tokens:"active,active"`
	Unknown  *int       `tokens:"unknown,,bad,bad"`
}

func tokenChart(t *testing.T) *sail.Navigator[tokenRoot] {
	t.Helper()
	c, err := sail.Compile[tokenRoot]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	return c
}

func checkTokenRecipe(t *testing.T, m *tokenRoot) {
	t.Helper()
	if m == nil || len(m.Items.Plain) != 4 || len(m.Items.Enum) != 1 || len(m.Items.Pair) != 1 || len(m.Items.Position) != 1 {
		t.Fatalf("selected occurrences: %#v", m)
	}
	p := m.Items.Plain
	if p[0].Name.Get() != "promoted" || !p[0].Active.Present() || !p[0].First.Present() || p[0].Second.Present() || p[0].First.Value.Get() != tokenSelf(17) || p[0].First.Optional.Present() {
		t.Fatalf("promoted first: %#v", p[0])
	}
	if p[1].Name.Get() != "other" || p[1].First.Present() || !p[1].Second.Present() || p[1].Second.Value.Get() != "other" || p[1].Active.Present() {
		t.Fatalf("second: %#v", p[1])
	}
	if p[2].First.Present() || p[2].Second.Present() || p[2].First.Value.Present() || p[2].Second.Value.Present() {
		t.Fatal("bare occurrence activated qualifier")
	}
	if !p[3].First.Present() || !p[3].Second.Present() || p[3].First.Value.Get() != tokenSelf(17) || p[3].Second.Value.Get() != tokenSelf(17) {
		t.Fatal("shared qualifier body lost occurrence values")
	}
	if m.Items.Enum[0].Name.Get() != "enum" || !m.Items.Enum[0].EnumOnly.Present() || m.Items.Pair[0].Name.Get() != "pair" || !m.Items.Position[0].Active.Present() {
		t.Fatal("discriminator or positional qualifier mismatch")
	}
	for i, want := range []reflect.Type{reflect.TypeFor[tokenSelf](), reflect.TypeFor[tokenOther](), nil, reflect.TypeFor[tokenSelf]()} {
		if want == nil {
			continue
		}
		v := p[i].First.Value
		if i == 1 {
			v = p[i].Second.Value
		}
		carrier, check := v.CarrierType()
		if !check.OK() || carrier.CapturedType() != want {
			t.Fatalf("occurrence %d carrier: %v", i, check)
		}
		result, check := v.Result()
		resultType := want
		if i == 1 {
			resultType = reflect.TypeFor[string]()
		}
		if !check.OK() || result.Type().CapturedType() != resultType {
			t.Fatalf("occurrence %d result: %v", i, check)
		}
	}
}

func TestPopulationTokensSelectionAndIsolation(t *testing.T) {
	c := tokenChart(t)
	m, err := c.Build[tokenRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	checkTokenRecipe(t, m)
	again, err := c.Populate(tokenRecipe{})
	if err != nil {
		t.Fatal(err)
	}
	checkTokenRecipe(t, again)
	if m == again {
		t.Fatal("Chart reused model")
	}
	for range 8 {
		t.Run("concurrent reuse", func(t *testing.T) {
			t.Parallel()
			got, err := c.Build[tokenRecipe]()
			if err != nil {
				t.Fatal(err)
			}
			checkTokenRecipe(t, got)
		})
	}
	checkTokenRecipe(t, m)
}

// Construct distinct Recipe types to vary only the raw field tag.
func tokenInput(tag string, carrier reflect.Type) any {
	t := reflect.StructOf([]reflect.StructField{{Name: "Item", Type: carrier, Tag: reflect.StructTag(`tokens:"` + tag + `"`)}})
	return reflect.New(t).Elem().Interface()
}

func TestPopulationTokensInvalidBeforeGetter(t *testing.T) {
	c := tokenChart(t)
	for _, tag := range []string{"flag,typo", "flag,", "flag,,first", "flag,first,first", "flag, first", "flag,first ", "flag,First", "flag,enumonly", "flag,lateonly"} {
		t.Run(tag, func(t *testing.T) {
			tokenCalls.Store(0)
			m, err := c.Populate(tokenInput(tag, reflect.TypeFor[tokenSelf]()))
			if err == nil || m != nil || !strings.Contains(err.Error(), "invalid_recipe_token") {
				t.Fatalf("tag %q: model=%v error=%v", tag, m, err)
			}
			if tokenCalls.Load() != 0 {
				t.Fatal("invalid token invoked getter")
			}
		})
	}
}

func TestPopulationTokensUnmatchedSelectorBeforeGetter(t *testing.T) {
	c := tokenChart(t)
	tokenCalls.Store(0)
	// Both discriminator words are known, but only {enum} and {enum,multi}
	// identify declarations. {multi} is a selector miss, not an unknown token.
	legend, err := c.Populate(tokenInput("flag,multi", reflect.TypeFor[tokenSelf]()))
	if err == nil || legend != nil || !strings.Contains(err.Error(), "does not match the remaining grammar") {
		t.Fatalf("unmatched selector: legend=%v error=%v", legend, err)
	}
	if tokenCalls.Load() != 0 {
		t.Fatal("unmatched selector invoked a getter")
	}
}

type tokenPlainOnlyLane struct {
	legend.Lane
	Plain []tokenPlain `sail:"spot" spot:"multi"`
}

type tokenEnumOnlyLane struct {
	legend.Lane
	Enum []tokenEnum `sail:"spot" spot:"multi,required"`
}

type tokenRetryRoot struct {
	legend.Legend `sail:"legend" mark:"tokens"`
	First         tokenPlainOnlyLane `sail:"lane"`
	Later         tokenEnumOnlyLane  `sail:"lane"`
}

type tokenRetryRecipe struct {
	Only bool `tokens:"flag,enum,enumonly" name:"retained"`
}

func TestPopulationTokensRetryClassifiedCandidateInLaterLane(t *testing.T) {
	chart, err := sail.Compile[tokenRetryRoot]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[tokenRetryRecipe]()
	if err != nil || populatedLegend == nil {
		t.Fatalf("later selector: legend=%v error=%v", populatedLegend, err)
	}
	if len(populatedLegend.First.Plain) != 0 || len(populatedLegend.Later.Enum) != 1 {
		t.Fatal("classified candidate was consumed by or lost in the first Lane")
	}
	if occurrence := populatedLegend.Later.Enum[0]; occurrence.Name.Get() != "retained" || !occurrence.EnumOnly.Present() {
		t.Fatal("the original candidate or its Qualifier was lost during Lane advancement")
	}
}

func TestPopulationTokensQualifierFailures(t *testing.T) {
	c := tokenChart(t)
	tokenCalls.Store(0)
	bare, err := c.Populate(tokenInput("flag", reflect.TypeFor[tokenSelf]()))
	if err != nil || bare == nil || len(bare.Items.Plain) != 1 || bare.Items.Plain[0].First.Value.Present() || tokenCalls.Load() != 0 {
		t.Fatalf("inactive: model=%v error=%v calls=%d", bare, err, tokenCalls.Load())
	}
	good, err := c.Populate(tokenInput("flag,first", reflect.TypeFor[tokenSelf]()))
	if err != nil || good == nil {
		t.Fatalf("good: %v", err)
	}
	for _, typ := range []reflect.Type{reflect.TypeFor[bool](), reflect.TypeFor[tokenPanic]()} {
		m, err := c.Populate(tokenInput("flag,first", typ))
		if err == nil || m != nil || !strings.Contains(err.Error(), "Item") {
			t.Fatalf("active %v: model=%v error=%v", typ, m, err)
		}
	}
	if good.Items.Plain[0].First.Value.Get() != tokenSelf(17) || !good.Items.Plain[0].First.Present() {
		t.Fatal("failure changed previous model")
	}
}

type tokenInner struct {
	legend.Region `sail:"region" family:"inner"`
}
type tokenInnerLane struct {
	legend.Lane
	Inner tokenInner `sail:"spot" spot:"single"`
}
type tokenOuterRoot struct {
	legend.Legend `sail:"legend" mark:"tokens"`
	Inner         tokenInnerLane `sail:"lane"`
	Items         tokenLane      `sail:"lane"`
}
type tokenOutOfScopeInner struct {
	Out     *int `tokens:"flag,,unknown,unknown"`
	Unknown *int `tokens:"unrecognized,,bad,bad"`
}
type tokenOutOfScopeRecipe struct {
	Inner tokenOutOfScopeInner `tokens:"inner"`
}

func TestPopulationTokensIgnoreOutOfScopeBeforeClassification(t *testing.T) {
	c, err := sail.Compile[tokenOuterRoot]()
	if err != nil {
		t.Fatal(err)
	}
	m, err := c.Build[tokenOutOfScopeRecipe]()
	if err != nil || m == nil {
		t.Fatalf("out-of-scope fields: model=%v error=%v", m, err)
	}
}
