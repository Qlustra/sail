package populating

import (
	"fmt"

	"github.com/qlustra/intel"
	"github.com/qlustra/intel/capture"
	"github.com/qlustra/sail/internal/chart/declarations"
)

type Populator[L any] struct {
	// Grammar is Sail's carrier graph; legendValue captures the Framework's Legend,
	// and chartType describes the App's declaration without retaining its value.
	Grammar     *declarations.Legend
	legendValue *capture.ValueInterrogator
	chartType   *capture.TypeInterrogator
}

func (p *Populator[L]) Populate(chartType *capture.TypeInterrogator) (*L, error) {
	p.chartType = chartType // Preserve exact type evidence even if admission fails.
	if p.Grammar == nil {
		return nil, fmt.Errorf("cannot populate without a compiled grammar")
	}
	if chartType == nil || chartType.TypeInspector == nil {
		return nil, fmt.Errorf("Chart type is missing")
	}
	if chartType.IsPointerKind() {
		return nil, fmt.Errorf("Chart nodes must be values, not pointers: %s", chartType)
	}
	chartStruct, check := chartType.AsStruct()
	if !check.OK() {
		return nil, fmt.Errorf("Chart root must be a struct value: %s", check.Why())
	}
	legend := new(L)
	legendValue, check := intel.CaptureValue(legend)
	if !check.OK() {
		return nil, fmt.Errorf("cannot capture model value: %s", check.Why())
	}
	p.legendValue = legendValue
	legendScope, check := legendValue.AsStruct()
	if !check.OK() {
		return nil, fmt.Errorf("Legend destination must be a struct: %s", check.Why())
	}
	if err := p.scope(chartStruct, nil, legendScope, p.Grammar.Lanes); err != nil {
		return nil, err
	}
	return legend, nil
}
