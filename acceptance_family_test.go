package sail_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
	"github.com/qlustra/sail/testdata/contractfamily"
)

type familyLandmark[Family legend.ContractFamily[Carrier], Carrier any] struct {
	legend.Landmark
}

func (familyLandmark[Family, Carrier]) Family() string { return "carrier" }
func (familyLandmark[Family, Carrier]) Accept() legend.Acceptance {
	// The compiler infers Carrier from Family's ContractOf parameter.
	return legend.AcceptContractFamily[Family]()
}

type familyLegend[Family legend.ContractFamily[Carrier], Carrier any] struct {
	legend.Legend
	Lane struct {
		legend.Lane
		Carriers legend.MultiSpot[familyLandmark[Family, Carrier]] `sail:"spot" multiplicity:"any"`
	} `sail:"lane"`
}

func (familyLegend[Family, Carrier]) Mark() string { return "acceptance" }

type familyValueWrapper struct{ contractfamily.Cell[int] }
type familyNestedWrapper struct{ familyValueWrapper }
type familyPointerWrapper struct{ *contractfamily.Cell[int] }

type familyExtra struct{ contractfamily.Cell[int] }

func (familyExtra) Extra() string { panic("Extra must not run during acceptance") }

type familyOverride struct{ contractfamily.Cell[int] }

func (familyOverride) Get() int { panic("override must not run during acceptance") }

type familyWrongGet struct{ contractfamily.Cell[int] }

func (familyWrongGet) Get() string { panic("wrong Get must not run during acceptance") }

type familyWrongSet struct{ contractfamily.Cell[int] }

func (*familyWrongSet) Set(string) { panic("wrong Set must not run during acceptance") }

type familyWrongReturn struct{ contractfamily.Cell[int] }

func (*familyWrongReturn) Set(int) error { panic("wrong Set must not run during acceptance") }

// Same spelling in another package cannot replace the package-owned witness.
func (*familyWrongSet) contract() reflect.Type { return reflect.TypeFor[any]() }

type foreignFamilyWitness[API any] struct{}

func (foreignFamilyWitness[API]) contract() reflect.Type { return reflect.TypeFor[API]() }

type familySpoofed struct {
	contractfamily.Cell[int]
	foreignFamilyWitness[any]
}

func (*familySpoofed) Set(string) { panic("spoofed Set must not run during acceptance") }

type familyWeak struct{}

func (familyWeak) Get() int                { panic("weak Get must not run during acceptance") }
func (*familyWeak) Set(int)                { panic("weak Set must not run during acceptance") }
func (*familyWeak) contract() reflect.Type { return reflect.TypeFor[any]() }

func TestAcceptContractFamilyUsesInstantiatedNativeInterface(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		requireFamilyAccepted[contractfamily.Cell[int], contractfamily.Family](t)
	})
	t.Run("string", func(t *testing.T) {
		requireFamilyAccepted[contractfamily.Cell[string], contractfamily.Family](t)
	})
	t.Run("value embedding", func(t *testing.T) {
		requireFamilyAccepted[familyValueWrapper, contractfamily.Family](t)
	})
	t.Run("nested embedding", func(t *testing.T) {
		requireFamilyAccepted[familyNestedWrapper, contractfamily.Family](t)
	})
	t.Run("pointer carrier field", func(t *testing.T) {
		requireFamilyAccepted[*familyValueWrapper, contractfamily.Family](t)
	})
	t.Run("pointer chain carrier field", func(t *testing.T) {
		requireFamilyAccepted[**familyValueWrapper, contractfamily.Family](t)
	})
	t.Run("additional methods", func(t *testing.T) {
		requireFamilyAccepted[familyExtra, contractfamily.Family](t)
	})
	t.Run("compatible override", func(t *testing.T) {
		requireFamilyAccepted[familyOverride, contractfamily.Family](t)
	})
	t.Run("incorrect typed result", func(t *testing.T) {
		requireFamilyRejected[familyWrongGet, contractfamily.Family](t, "expected contract")
	})
	t.Run("incorrect typed argument and foreign private override", func(t *testing.T) {
		requireFamilyRejected[familyWrongSet, contractfamily.Family](t, "expected contract")
	})
	t.Run("incorrect result count", func(t *testing.T) {
		requireFamilyRejected[familyWrongReturn, contractfamily.Family](t, "expected contract")
	})
	t.Run("foreign embedded witness cannot weaken the API", func(t *testing.T) {
		requireFamilyRejected[familySpoofed, contractfamily.Family](t, "expected contract")
	})
	t.Run("foreign private witness is not a family provider", func(t *testing.T) {
		requireFamilyRejected[familyWeak, contractfamily.Family](t, "requires provider")
	})
	t.Run("interface declaration cannot supply a surrogate", func(t *testing.T) {
		requireFamilyRejected[contractfamily.API[int], contractfamily.Family](t, "cannot supply a concrete zero-valued witness")
	})
	t.Run("internal embedded pointer is not initialized", func(t *testing.T) {
		requireFamilyRejected[familyPointerWrapper, contractfamily.Family](t, "panicked on its zero-valued receiver")
	})
}

func TestAcceptContractFamilyUsesFreshZeroSurrogates(t *testing.T) {
	charter, err := sail.CompileLegend[familyLegend[contractfamily.Family, contractfamily.Provider]]()
	if err != nil {
		t.Fatal(err)
	}
	type twice struct {
		First  familyValueWrapper `acceptance:"carrier"`
		Second familyValueWrapper `acceptance:"carrier"`
	}
	for range 2 {
		chart, err := sail.Populate[twice](charter)
		if err != nil {
			t.Fatalf("fresh surrogate population: %v", err)
		}
		live := &twice{}
		live.First.WitnessVisits = 41
		live.Second.WitnessVisits = 42
		if _, err := sail.Materialize(charter, chart, live); err != nil {
			t.Fatal(err)
		}
		if live.First.WitnessVisits != 41 || live.Second.WitnessVisits != 42 {
			t.Fatal("family metadata changed the live Subject")
		}
	}
}

type nilContractFamily struct{}

func (nilContractFamily) ContractOf(contractfamily.Provider) reflect.Type { return nil }

type concreteContractFamily struct{}

func (concreteContractFamily) ContractOf(contractfamily.Provider) reflect.Type {
	return reflect.TypeFor[contractfamily.Cell[int]]()
}

type selfContractFamily struct{}

func (selfContractFamily) ContractOf(contractfamily.Provider) reflect.Type {
	return reflect.TypeFor[acceptanceSelfAPI]()
}

type panicContractFamily struct{}

func (panicContractFamily) ContractOf(contractfamily.Provider) reflect.Type {
	panic("family metadata failed")
}

func TestAcceptContractFamilyRejectsInvalidResolvedContracts(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		requireFamilyRejected[familyValueWrapper, nilContractFamily](t, "contract type must not be nil")
	})
	t.Run("concrete type", func(t *testing.T) {
		requireFamilyRejected[familyValueWrapper, concreteContractFamily](t, "must be an interface")
	})
	t.Run("Self template", func(t *testing.T) {
		requireFamilyRejected[familyValueWrapper, selfContractFamily](t, "Self is unsupported")
	})
	t.Run("descriptor panic", func(t *testing.T) {
		requireFamilyRejected[familyValueWrapper, panicContractFamily](t, "family metadata failed")
	})
}

type pointerContractFamily struct{}

func (*pointerContractFamily) ContractOf(contractfamily.Provider) reflect.Type {
	return reflect.TypeFor[contractfamily.API[int]]()
}

type interfaceContractFamily interface {
	ContractOf(contractfamily.Provider) reflect.Type
}

type scalarContractFamily int

func (scalarContractFamily) ContractOf(contractfamily.Provider) reflect.Type {
	return reflect.TypeFor[contractfamily.API[int]]()
}

type concreteProviderFamily struct{}

func (concreteProviderFamily) ContractOf(*contractfamily.Cell[int]) reflect.Type {
	return reflect.TypeFor[contractfamily.API[int]]()
}

type selfProviderFamily struct{}

func (selfProviderFamily) ContractOf(acceptanceSelfAPI) reflect.Type {
	return reflect.TypeFor[any]()
}

func TestAcceptContractFamilyRejectsInvalidDeclarations(t *testing.T) {
	t.Run("pointer descriptor", func(t *testing.T) {
		requireInvalidFamily[*pointerContractFamily](t, "concrete non-pointer struct")
	})
	t.Run("interface descriptor", func(t *testing.T) {
		requireInvalidFamily[interfaceContractFamily](t, "concrete non-pointer struct")
	})
	t.Run("scalar descriptor", func(t *testing.T) {
		requireInvalidFamily[scalarContractFamily](t, "concrete non-pointer struct")
	})
	t.Run("concrete provider", func(t *testing.T) {
		requireInvalidFamily[concreteProviderFamily](t, "must be an interface")
	})
	t.Run("Self provider", func(t *testing.T) {
		requireInvalidFamily[selfProviderFamily](t, "Self is unsupported")
	})
}

func requireFamilyAccepted[Subject any, Family legend.ContractFamily[Carrier], Carrier any](t *testing.T) {
	t.Helper()
	charter, err := sail.CompileLegend[familyLegend[Family, Carrier]]()
	if err != nil {
		t.Fatalf("compile family: %v", err)
	}
	if _, err := sail.Populate[acceptanceSubject[Subject]](charter); err != nil {
		t.Fatalf("compatible family carrier rejected: %v", err)
	}
}

func requireFamilyRejected[Subject any, Family legend.ContractFamily[Carrier], Carrier any](t *testing.T, message string) {
	t.Helper()
	charter, err := sail.CompileLegend[familyLegend[Family, Carrier]]()
	if err != nil {
		t.Fatalf("compile family: %v", err)
	}
	_, err = sail.Populate[acceptanceSubject[Subject]](charter)
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected rejection containing %q, got %v", message, err)
	}
}

func requireInvalidFamily[Family legend.ContractFamily[Carrier], Carrier any](t *testing.T, message string) {
	t.Helper()
	_, err := sail.CompileLegend[familyLegend[Family, Carrier]]()
	if err == nil || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected compilation error containing %q, got %v", message, err)
	}
}
