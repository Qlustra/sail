package populating

import (
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
)

func TestPopulateRetainsExactChartType(t *testing.T) {
	type recipe struct{}
	typ := capture.Type[*recipe]()
	p := &Populator[struct{}]{Grammar: &declarations.Legend{}}
	legend, err := p.Populate(typ)
	if err == nil || legend != nil {
		t.Fatalf("legend=%v error=%v", legend, err)
	}
	if p.chartType != typ || p.chartType.CapturedType() != capture.Type[*recipe]().CapturedType() {
		t.Fatal("Chart type pointer or pointer type was not retained")
	}
	if p.legendValue != nil {
		t.Fatal("rejected Chart allocated a model")
	}
}
