package legend

import "github.com/qlustra/sail/legend"

type SailLegend struct {
	legend.Legend `mark:"sail"`
	Lanes         SailLane `sail:"lane"`
}

type SailLane struct {
	legend.Lane `family:"lane"`
}
