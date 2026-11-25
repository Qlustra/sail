package definitions

import (
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/legend"
)

// The helper also satisfies behavior.Region through promotion, but it is
// authored grammar, not the exact library-owned legend.Region boundary.
type boundaryRegionHelper struct {
	legend.Region
	Name legend.Feature `sail:"feature" symbol:"name"`
}

type boundaryAuthoredRegion struct {
	boundaryRegionHelper `sail:"region" family:"command"`
	Help                 legend.Feature `sail:"feature" symbol:"help"`
}

type boundaryPrivateHelper struct {
	legend.Region
	private string
}

type boundaryPrivateRegion struct {
	boundaryPrivateHelper `sail:"region" family:"command"`
}

type boundaryBadMarkerRegion struct {
	boundaryRegionHelper `sail:"region"` // The marker itself still needs a family.
}

type boundaryQualifierShadow struct {
	legend.Qualifier `sail:"qualifier"`
	present          bool                                   // Authored state, not the marker-owned promoted field.
	Value            legend.Barrel[string, countedProvider] `sail:"barrel" populate:"Value"`
}

func TestMarkerBoundaryKeepsAuthoredHelperMembers(t *testing.T) {
	definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	regionType, check := capture.Type[boundaryAuthoredRegion]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	context, validation := definition.Compile(regionType, nil)
	if !validation.IsOkay() {
		t.Fatal(validation.Trace())
	}
	regionDeclaration, check := context.Access[declarations.Region](define.DeclarationKey)
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if regionDeclaration.Family != "command" || len(regionDeclaration.Features) != 2 {
		t.Fatalf("authored helper members lost: family=%q features=%v", regionDeclaration.Family, regionDeclaration.Features)
	}
	for _, symbol := range []string{"name", "help"} {
		found := false
		for _, featureCarrier := range regionDeclaration.Features {
			found = found || featureCarrier.Symbol == symbol
		}
		if !found {
			t.Fatalf("missing authored Feature %q", symbol)
		}
	}
}

func TestMarkerBoundaryDoesNotIgnoreAuthoredErrors(t *testing.T) {
	for _, regionType := range []*capture.TypeInterrogator{
		capture.Type[boundaryPrivateRegion](),
		capture.Type[boundaryBadMarkerRegion](),
	} {
		definition, err := define.Struct[RegionDefinition, declarations.Region](nil)
		if err != nil {
			t.Fatal(err)
		}
		regionStruct, check := regionType.AsStruct()
		if !check.OK() {
			t.Fatal(check.Why())
		}
		if _, validation := definition.Compile(regionStruct, nil); validation.IsOkay() {
			t.Fatalf("invalid authored body passed: %s", regionType)
		}
	}
	definition, err := define.Struct[QualifierDefinition, declarations.Qualifier](nil)
	if err != nil {
		t.Fatal(err)
	}
	qualifierType, check := capture.Type[boundaryQualifierShadow]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	if _, validation := definition.Compile(qualifierType, nil); validation.IsOkay() {
		t.Fatal("authored private field shadowing marker state was ignored")
	}
}
