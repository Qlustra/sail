package sail

import (
	"sync"
	"testing"

	"github.com/qlustra/sail/internal/chart/declarations"
)

func TestNavigatorWarnings(t *testing.T) {
	var nilNavigator *Navigator[struct{}]
	if nilNavigator.Warnings() != nil || (&Navigator[struct{}]{}).Warnings() != nil {
		t.Fatal("nil and zero Navigators must have nil warnings")
	}
	want := Warning{Code: "example", Message: "explanation"}
	navigator := &Navigator[struct{}]{grammar: &declarations.Legend{Warnings: []declarations.Warning{want}}}
	first := navigator.Warnings()
	if len(first) != 1 || first[0] != want {
		t.Fatalf("warnings = %#v, want %#v", first, want)
	}
	first[0] = Warning{Code: "changed"}
	if got := navigator.Warnings(); len(got) != 1 || got[0] != want {
		t.Fatalf("mutating snapshot changed Navigator warnings: %#v", got)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got := navigator.Warnings()
			if len(got) != 1 || got[0] != want {
				t.Errorf("concurrent warnings = %#v, want %#v", got, want)
				return
			}
			got[0].Message = "local change"
		}()
	}
	wg.Wait()
}
