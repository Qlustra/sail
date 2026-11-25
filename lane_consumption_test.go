package sail_test

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type consumptionOptionalFlag struct {
	legend.Lane
	Flag orderBool `sail:"spot" spot:"single"`
}

type consumptionBoundaryAndFlag struct {
	legend.Lane
	Boundary orderSeparator `sail:"spot" spot:"single"`
	Flag     orderBool      `sail:"spot" spot:"single,required"`
}

type consumptionDependent struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag    `sail:"lane" exploration:"eager"`
	Later         consumptionBoundaryAndFlag `sail:"lane"`
}

type consumptionRequiredSingle struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Later         orderBoolLane           `sail:"lane"`
}

type consumptionMultiLane struct {
	legend.Lane
	Flags []orderBool `sail:"spot" spot:"multi,required"`
}

type consumptionRequiredMulti struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Later         consumptionMultiLane    `sail:"lane"`
}

type consumptionOptionalMulti struct {
	legend.Lane
	Flags []orderBool `sail:"spot" spot:"multi"`
}

type consumptionStarvedMulti struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalMulti `sail:"lane" exploration:"eager"`
	Later         consumptionMultiLane     `sail:"lane"`
}

type consumptionEmptyLane struct{ legend.Lane }
type consumptionEmptyBridge struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Bridge        consumptionEmptyLane    `sail:"lane"`
	Later         orderBoolLane           `sail:"lane"`
}
type consumptionCoveredBridge struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Bridge        consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Later         orderBoolLane           `sail:"lane"`
}
type consumptionOptionalSeparator struct {
	legend.Lane
	Boundary orderSeparator `sail:"spot" spot:"single"`
}
type consumptionBoundaryBridge struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag      `sail:"lane" exploration:"eager"`
	Bridge        consumptionOptionalSeparator `sail:"lane"`
	Later         orderBoolLane                `sail:"lane"`
}
type consumptionRequiredBoundary struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Boundary      orderSeparatorLane      `sail:"lane"`
	Later         orderBoolLane           `sail:"lane"`
}
type consumptionOptionalOverlap struct {
	legend.Legend `sail:"legend" mark:"order"`
	Eager         consumptionOptionalFlag `sail:"lane" exploration:"eager"`
	Later         consumptionOptionalFlag `sail:"lane"`
}

type consumptionRecipe struct {
	First    bool `order:"flag" name:"first"`
	Boundary bool `order:"separator"`
	Last     bool `order:"flag" name:"last"`
}
type consumptionMissing struct {
	First bool `order:"flag" name:"first"`
}

func consumptionWarning(t *testing.T, warnings []sail.Warning, names ...string) {
	t.Helper()
	for _, w := range warnings {
		if w.Code != "eager_lane_entry_boundary" {
			continue
		}
		message := strings.ToLower(w.Message)
		for _, name := range names {
			if !strings.Contains(message, strings.ToLower(name)) {
				t.Errorf("warning %q does not identify %q", w.Message, name)
			}
		}
		return
	}
	t.Errorf("missing eager_lane_entry_boundary warning: %#v", warnings)
}

func TestLaneConsumptionBoundaryDependentPopulation(t *testing.T) {
	c, err := sail.Compile[consumptionDependent]()
	if err != nil || c == nil {
		t.Fatalf("Compile: chart=%v error=%v", c, err)
	}
	want := c.Warnings()
	consumptionWarning(t, want, "Eager", "Later", "boundary")
	m, err := c.Build[consumptionRecipe]()
	if err != nil || m == nil {
		t.Fatalf("Build: model=%v error=%v", m, err)
	}
	for _, tc := range []struct {
		f    legend.Feature
		name string
	}{{m.Eager.Flag.Name, "first"}, {m.Later.Flag.Name, "last"}} {
		if !tc.f.Present() || tc.f.Get() != tc.name {
			t.Errorf("name=(%q,%v), want %q", tc.f.Get(), tc.f.Present(), tc.name)
		}
	}
	if got := c.Warnings(); !reflect.DeepEqual(got, want) {
		t.Errorf("warnings changed after success: %#v != %#v", got, want)
	}
	if m, err := c.Build[consumptionMissing](); err == nil || m != nil {
		t.Errorf("missing required later flag: model=%v error=%v", m, err)
	}
	if got := c.Warnings(); !reflect.DeepEqual(got, want) {
		t.Errorf("warnings changed after failure: %#v != %#v", got, want)
	}
	if len(want) > 0 {
		want[0].Message = "mutated snapshot"
		if reflect.DeepEqual(c.Warnings(), want) {
			t.Error("warning snapshot aliases Chart")
		}
	}
}

func TestLaneConsumptionUnreachableRequired(t *testing.T) {
	tests := []struct {
		name    string
		compile func() (bool, error)
	}{
		{"single", func() (bool, error) { c, e := sail.Compile[consumptionRequiredSingle](); return c == nil, e }},
		{"multi", func() (bool, error) { c, e := sail.Compile[consumptionRequiredMulti](); return c == nil, e }},
		{"greedy multi before required multi", func() (bool, error) { c, e := sail.Compile[consumptionStarvedMulti](); return c == nil, e }},
		{"empty bridge", func() (bool, error) { c, e := sail.Compile[consumptionEmptyBridge](); return c == nil, e }},
		{"covered bridge", func() (bool, error) { c, e := sail.Compile[consumptionCoveredBridge](); return c == nil, e }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nilChart, err := tc.compile()
			if !nilChart || err == nil {
				t.Fatalf("expected nil Chart and error: nil=%v error=%v", nilChart, err)
			}
			message := strings.ToLower(err.Error())
			for _, word := range []string{"eager", "later", "unreachable", "required"} {
				if !strings.Contains(message, word) {
					t.Errorf("error %q missing %q", err, word)
				}
			}
		})
	}
}

func TestLaneConsumptionInterveningBoundaries(t *testing.T) {
	c, err := sail.Compile[consumptionBoundaryBridge]()
	if err != nil || c == nil {
		t.Fatalf("optional boundary: chart=%v error=%v", c, err)
	}
	consumptionWarning(t, c.Warnings(), "Eager", "Later", "boundary")
	required, err := sail.Compile[consumptionRequiredBoundary]()
	if err != nil || required == nil {
		t.Fatalf("required boundary: chart=%v error=%v", required, err)
	}
	for _, w := range required.Warnings() {
		if w.Code == "eager_lane_entry_boundary" {
			t.Errorf("required boundary should eliminate advisory: %#v", w)
		}
	}
	if m, err := required.Build[consumptionRecipe](); m == nil || err != nil {
		t.Errorf("required boundary recipe: model=%v error=%v", m, err)
	}
	optional, err := sail.Compile[consumptionOptionalOverlap]()
	if err != nil || optional == nil {
		t.Fatalf("optional overlap: chart=%v error=%v", optional, err)
	}
	for _, w := range optional.Warnings() {
		if w.Code == "eager_lane_entry_boundary" {
			t.Errorf("optional overlap advisory: %#v", w)
		}
	}
}

// Capture both streams concurrently so even unexpected voluminous output cannot fill a pipe.
// This test and its subtests must remain nonparallel while process globals are redirected.
func TestLaneConsumptionCompileIsSilent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		compile func() (bool, error)
	}{
		{"warning", func() (bool, error) { c, e := sail.Compile[consumptionDependent](); return c != nil, e }},
		{"error", func() (bool, error) { c, e := sail.Compile[consumptionRequiredSingle](); return c == nil, e }},
		{"body error", func() (bool, error) { c, e := sail.Compile[struct{}](); return c == nil, e }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outR, outW, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			errR, errW, err := os.Pipe()
			if err != nil {
				outR.Close()
				outW.Close()
				t.Fatal(err)
			}
			oldOut, oldErr := os.Stdout, os.Stderr
			os.Stdout, os.Stderr = outW, errW
			type result struct {
				data []byte
				err  error
			}
			read := func(r *os.File, ch chan<- result) {
				var b bytes.Buffer
				_, e := io.Copy(&b, r)
				r.Close()
				ch <- result{b.Bytes(), e}
			}
			outCh, errCh := make(chan result, 1), make(chan result, 1)
			go read(outR, outCh)
			go read(errR, errCh)
			var ok bool
			var compileErr error
			func() {
				defer func() { os.Stdout, os.Stderr = oldOut, oldErr; outW.Close(); errW.Close() }()
				ok, compileErr = tc.compile()
			}()
			out, stderr := <-outCh, <-errCh
			if out.err != nil || stderr.err != nil {
				t.Fatalf("read pipes: %v, %v", out.err, stderr.err)
			}
			if !ok || (tc.name == "warning" && compileErr != nil) || (tc.name != "warning" && compileErr == nil) {
				t.Errorf("compile outcome: ok=%v error=%v", ok, compileErr)
			}
			if len(out.data) != 0 || len(stderr.data) != 0 {
				t.Errorf("unexpected output: stdout=%q stderr=%q", out.data, stderr.data)
			}
		})
	}
}
