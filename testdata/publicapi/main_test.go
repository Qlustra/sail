package main

import (
	"bytes"
	"slices"
	"testing"

	"github.com/qlustra/sail"
)

func TestQuickStart(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "compile\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestLocationIsNameableAndResolves(t *testing.T) {
	navigator, err := sail.Compile[PipelineLegend]()
	if err != nil {
		t.Fatal(err)
	}
	model, err := navigator.Build[ReleaseChart]()
	if err != nil {
		t.Fatal(err)
	}
	location, check := model.Jobs.Items[0].ChartLocation()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	var publicLocation *sail.ChartLocation = location
	if got := publicLocation.Index(); !slices.Equal(got, []int{0}) {
		t.Fatalf("index = %v, want [0]", got)
	}
	if _, check := publicLocation.Resolve(&ReleaseChart{}); !check.OK() {
		t.Fatal(check.Why())
	}
}
