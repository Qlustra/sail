package definitions

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/define"
	"github.com/qlustra/sail/internal/chart/declarations"
	"github.com/qlustra/sail/legend"
)

type regionMissing struct {
	legend.Region `sail:"region"`
}
type regionEmpty struct {
	legend.Region `sail:"region" family:""`
}
type regionValid struct {
	legend.Region `sail:"region" family:"command"`
}
type regionOpaque struct {
	legend.Region `sail:"region" family:" Some role! "`
}
type regionComma struct {
	legend.Region `sail:"region" family:"a,b"`
}

func (regionMissing) Family() string { return "legacy" }
func (regionValid) Family() string   { return "legacy" }

type landmarkMissing struct {
	legend.Landmark `sail:"landmark"`
}
type landmarkEmpty struct {
	legend.Landmark `sail:"landmark" family:""`
}
type landmarkValid struct {
	legend.Landmark `sail:"landmark" family:"flag"`
}
type landmarkOther struct {
	legend.Landmark `sail:"landmark" family:"flag"`
}
type landmarkComma struct {
	legend.Landmark `sail:"landmark" family:"a,b"`
}

func (landmarkMissing) Family() string { return "legacy" }
func (landmarkValid) Family() string   { return "legacy" }

type sharedFamilyLane struct {
	legend.Lane
	First  landmarkValid `sail:"spot" spot:"single"`
	Second landmarkOther `sail:"spot" spot:"single"`
}

func TestFamilyCompilation(t *testing.T) {
	region, err := define.Struct[RegionDefinition, declarations.Region](nil)
	if err != nil {
		t.Fatal(err)
	}
	landmark, err := define.Struct[LandmarkDefinition, declarations.Landmark](nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		typ    reflect.Type
		region bool
		want   string
	}{
		{"region missing", reflect.TypeFor[regionMissing](), true, ""},
		{"region empty", reflect.TypeFor[regionEmpty](), true, ""},
		{"region comma", reflect.TypeFor[regionComma](), true, ""},
		{"region valid", reflect.TypeFor[regionValid](), true, "command"},
		{"region opaque value", reflect.TypeFor[regionOpaque](), true, " Some role! "},
		{"landmark missing", reflect.TypeFor[landmarkMissing](), false, ""},
		{"landmark empty", reflect.TypeFor[landmarkEmpty](), false, ""},
		{"landmark comma", reflect.TypeFor[landmarkComma](), false, ""},
		{"landmark valid", reflect.TypeFor[landmarkValid](), false, "flag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			subject, check := capture.TypeReflection(tc.typ).AsStruct()
			if !check.OK() {
				t.Fatal(check.Why())
			}
			if tc.region {
				context, validation := region.Compile(subject, nil)
				if validation.IsOkay() != (tc.want != "") {
					t.Fatalf("validation: %s", validation.Reason())
				}
				if tc.want != "" {
					declaration, check := context.Access[declarations.Region](define.DeclarationKey)
					if !check.OK() {
						t.Fatal(check.Why())
					}
					if declaration.Family != tc.want {
						t.Fatalf("family = %q, want %q", declaration.Family, tc.want)
					}
					if !strings.Contains(declaration.Trace(), "family="+strconv.Quote(tc.want)) {
						t.Fatal(declaration.Trace())
					}
				}
			} else {
				context, validation := landmark.Compile(subject, nil)
				if validation.IsOkay() != (tc.want != "") {
					t.Fatalf("validation: %s", validation.Reason())
				}
				if tc.want != "" {
					declaration, check := context.Access[declarations.Landmark](define.DeclarationKey)
					if !check.OK() {
						t.Fatal(check.Why())
					}
					if declaration.Family != tc.want {
						t.Fatalf("family = %q, want %q", declaration.Family, tc.want)
					}
					if !strings.Contains(declaration.Trace(), `family="flag"`) {
						t.Fatal(declaration.Trace())
					}
				}
			}
		})
	}
}

func TestSharedFamilyCollision(t *testing.T) {
	lane, err := define.Struct[LaneDefinition, declarations.Lane](nil)
	if err != nil {
		t.Fatal(err)
	}
	subject, check := capture.Type[sharedFamilyLane]().AsStruct()
	if !check.OK() {
		t.Fatal(check.Why())
	}
	_, validation := lane.Compile(subject, nil)
	if validation.IsOkay() || !strings.Contains(validation.Trace(), "duplicate_spot_selector") || !strings.Contains(validation.Trace(), `fields "First" and "Second"`) {
		t.Fatal(validation.Trace())
	}
}
