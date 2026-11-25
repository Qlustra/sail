package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type integrationLeaf struct {
	legend.Landmark `sail:"landmark" family:"flag"`
}
type integrationLane struct {
	legend.Lane
	Flag integrationLeaf `sail:"spot" spot:"single"`
}
type integrationInvalid struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         integrationLane `sail:"lane"`
	Second        integrationLane `sail:"lane" exploration:"eager"`
}
type integrationValid struct {
	legend.Legend `sail:"legend" mark:"x"`
	First         integrationLane `sail:"lane" exploration:"eager"`
	Second        integrationLane `sail:"lane"`
}

func TestCompileChecksCompletedLaneGraph(t *testing.T) {
	chart, err := sail.Compile[integrationInvalid]()
	if err == nil || chart != nil || !strings.Contains(err.Error(), `First`) || !strings.Contains(err.Error(), `Second`) {
		t.Fatalf("invalid graph: chart=%v error=%v", chart, err)
	}
	validChart, err := sail.Compile[integrationValid]()
	if err != nil || validChart == nil {
		t.Fatalf("valid graph: chart=%v error=%v", validChart, err)
	}
}
