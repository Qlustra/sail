package sail_test

import (
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type publicDiscriminatorRegion struct {
	legend.Region `sail:"region" family:"public" discriminate:"selection"`
	Presence      legend.Qualifier `sail:"qualifier" symbol:"presence"`
}

type publicDiscriminatorCollision struct {
	legend.Legend `sail:"legend" mark:"public"`
	Lane          publicCollisionLane `sail:"lane"`
}

type publicCollisionLane struct {
	legend.Lane
	Region publicDiscriminatorRegion `sail:"spot" spot:"single"`
	Leaf   publicDiscriminatorLeaf   `sail:"spot" spot:"single"`
}

type publicDiscriminatorLeaf struct {
	legend.Landmark `sail:"landmark" family:"public" discriminate:"other"`
	Presence        legend.Qualifier `sail:"qualifier" symbol:"selection"`
}

func TestCompileRejectsQualifierDiscriminatorGraphCollision(t *testing.T) {
	chart, err := sail.Compile[publicDiscriminatorCollision]()
	if err == nil || chart != nil {
		t.Fatalf("expected rejected graph and nil Chart; got chart=%v error=%v", chart, err)
	}
	for _, fragment := range []string{"public", "selection", "Presence", "landmark"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("error %q missing %q", err, fragment)
		}
	}
}
