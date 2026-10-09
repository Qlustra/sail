package sail_test

import (
	"testing"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type acceptanceLandmark[Contract any] struct {
	legend.Landmark
}

func (acceptanceLandmark[Contract]) Family() string { return "carrier" }
func (acceptanceLandmark[Contract]) Accept() legend.Acceptance {
	return legend.AcceptContract[Contract]()
}

type acceptanceLegend[Contract any] struct {
	legend.Legend
	Lane struct {
		legend.Lane
		Carriers legend.MultiSpot[acceptanceLandmark[Contract]] `sail:"spot" multiplicity:"any"`
	} `sail:"lane"`
}

func (acceptanceLegend[Contract]) Mark() string { return "acceptance" }

type acceptanceSubject[Carrier any] struct {
	Item Carrier `acceptance:"carrier"`
}

// Every method panics to verify that checking the public API never invokes it.
type acceptanceFile struct{}

func (acceptanceFile) Path() string                     { panic("Path invoked") }
func (acceptanceFile) ReadBytes() ([]byte, error)       { panic("ReadBytes invoked") }
func (acceptanceFile) AppendBytes([]byte, ...int) error { panic("AppendBytes invoked") }
func (acceptanceFile) CopyTo(acceptanceFile) error      { panic("CopyTo invoked") }
func (*acceptanceFile) ComposePath(string)              { panic("ComposePath invoked") }
func (acceptanceFile) privateFileMethod()               { panic("privateFileMethod invoked") }

type acceptanceFileWrapper struct{ acceptanceFile }
type acceptanceFilePointerWrapper struct{ *acceptanceFile }
type acceptanceFileNestedWrapper struct{ acceptanceFileWrapper }
type acceptanceFileDefined acceptanceFile

type acceptanceFileExtra struct{ acceptanceFile }

func (acceptanceFileExtra) Checksum() string { panic("Checksum invoked") }

type acceptanceFileOverride struct{ acceptanceFile }

func (acceptanceFileOverride) ReadBytes() ([]byte, error) { panic("override invoked") }

type acceptanceFileAPI interface {
	Path() string
	ReadBytes() ([]byte, error)
	AppendBytes([]byte, ...int) error
	CopyTo(acceptanceFile) error
	ComposePath(string)
}

type acceptanceFileExtraAPI interface {
	acceptanceFileAPI
	Checksum() string
}

type acceptanceFileValueAPI interface {
	Path() string
	ReadBytes() ([]byte, error)
	AppendBytes([]byte, ...int) error
	CopyTo(acceptanceFile) error
}

type acceptanceFileWeakAPI interface {
	Path() string
	ReadBytes() ([]byte, error)
}

type acceptanceFileWeak struct{}

func (acceptanceFileWeak) Path() string               { panic("weak Path invoked") }
func (acceptanceFileWeak) ReadBytes() ([]byte, error) { panic("weak ReadBytes invoked") }

type acceptanceFileWrongRead struct{ acceptanceFile }

func (acceptanceFileWrongRead) ReadBytes() string { panic("wrong ReadBytes invoked") }

type acceptanceFileWrongVariadic struct{ acceptanceFile }

func (acceptanceFileWrongVariadic) AppendBytes([]byte, []int) error {
	panic("wrong AppendBytes invoked")
}

type acceptanceFileWrongLiteral struct{ acceptanceFile }

func (acceptanceFileWrongLiteral) CopyTo(acceptanceFileWrongLiteral) error {
	panic("wrong CopyTo invoked")
}

type acceptanceFileShadow struct {
	acceptanceFile
	ReadBytes bool
}

type acceptanceFilePointerShadow struct {
	acceptanceFile
	ComposePath bool
}

type acceptanceOtherRead struct{}

func (acceptanceOtherRead) ReadBytes() ([]byte, error) { panic("other ReadBytes invoked") }

type acceptanceFileAmbiguous struct {
	acceptanceFile
	acceptanceOtherRead
}

func TestAcceptContractConcreteRequiresCompletePublicMethodSet(t *testing.T) {
	t.Run("original", func(t *testing.T) {
		requireContractAcceptance[acceptanceFile, acceptanceFile](t, true)
	})
	t.Run("original pointer", func(t *testing.T) {
		requireContractAcceptance[*acceptanceFile, acceptanceFile](t, true)
	})
	t.Run("embedded value", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWrapper, acceptanceFile](t, true)
	})
	t.Run("embedded pointer", func(t *testing.T) {
		requireContractAcceptance[acceptanceFilePointerWrapper, acceptanceFile](t, true)
	})
	t.Run("pointer to wrapper", func(t *testing.T) {
		requireContractAcceptance[*acceptanceFileWrapper, acceptanceFile](t, true)
	})
	t.Run("pointer chain to nested wrapper", func(t *testing.T) {
		requireContractAcceptance[**acceptanceFileNestedWrapper, acceptanceFile](t, true)
	})
	t.Run("additional methods", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileExtra, acceptanceFile](t, true)
	})
	t.Run("compatible override", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileOverride, acceptanceFile](t, true)
	})
	t.Run("complete interface", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileAPI, acceptanceFile](t, true)
	})
	t.Run("interface with additional methods", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileExtraAPI, acceptanceFile](t, true)
	})
	t.Run("pointer donor", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWrapper, *acceptanceFile](t, true)
	})
	t.Run("pointer chain donor", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileAPI, **acceptanceFile](t, true)
	})
	t.Run("weak carrier", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWeak, acceptanceFile](t, false)
	})
	t.Run("weak interface", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWeakAPI, acceptanceFile](t, false)
	})
	t.Run("all value methods without pointer method", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileValueAPI, acceptanceFile](t, false)
	})
	t.Run("incompatible override", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWrongRead, acceptanceFile](t, false)
	})
	t.Run("variadicness remains part of signature", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWrongVariadic, acceptanceFile](t, false)
	})
	t.Run("concrete arguments do not become Self", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWrongLiteral, acceptanceFile](t, false)
	})
	t.Run("field shadows method", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileShadow, acceptanceFile](t, false)
	})
	t.Run("field shadows pointer method", func(t *testing.T) {
		requireContractAcceptance[acceptanceFilePointerShadow, acceptanceFile](t, false)
	})
	t.Run("ambiguous promotion", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileAmbiguous, acceptanceFile](t, false)
	})
	t.Run("defined type does not inherit methods", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileDefined, acceptanceFile](t, false)
	})
}

type acceptancePrivateAPI interface {
	Path() string
	privateAcceptance()
}

type acceptancePrivateCarrier struct{}

func (acceptancePrivateCarrier) Path() string       { panic("private Path invoked") }
func (acceptancePrivateCarrier) privateAcceptance() { panic("private marker invoked") }

type acceptancePrivateWrapper struct{ acceptancePrivateCarrier }

type acceptanceSelfAPI interface {
	Clone() legend.Self
	Same(legend.Self) bool
}

type acceptanceSelfCarrier struct{}

func (acceptanceSelfCarrier) Clone() acceptanceSelfCarrier    { panic("Clone invoked") }
func (acceptanceSelfCarrier) Same(acceptanceSelfCarrier) bool { panic("Same invoked") }

type acceptanceSelfWrong struct{}

func (acceptanceSelfWrong) Clone() *acceptanceSelfWrong   { panic("wrong Clone invoked") }
func (acceptanceSelfWrong) Same(acceptanceSelfWrong) bool { panic("wrong Same invoked") }

type acceptanceSelfInterface interface {
	Clone() acceptanceSelfInterface
	Same(acceptanceSelfInterface) bool
}

func TestAcceptContractPreservesInterfaceAndSelfMatching(t *testing.T) {
	t.Run("ordinary interface", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWrapper, acceptanceFileAPI](t, true)
	})
	t.Run("ordinary interface still rejects missing methods", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWeak, acceptanceFileAPI](t, false)
	})
	t.Run("private interface marker", func(t *testing.T) {
		requireContractAcceptance[acceptancePrivateWrapper, acceptancePrivateAPI](t, true)
	})
	t.Run("private interface marker remains required", func(t *testing.T) {
		requireContractAcceptance[acceptanceFileWeak, acceptancePrivateAPI](t, false)
	})
	t.Run("Self uses resolved carrier type", func(t *testing.T) {
		requireContractAcceptance[**acceptanceSelfCarrier, acceptanceSelfAPI](t, true)
	})
	t.Run("Self rejects pointer result for value carrier", func(t *testing.T) {
		requireContractAcceptance[acceptanceSelfWrong, acceptanceSelfAPI](t, false)
	})
	t.Run("Self supports interface carrier", func(t *testing.T) {
		requireContractAcceptance[acceptanceSelfInterface, acceptanceSelfAPI](t, true)
	})
}

func TestAcceptContractRejectsNonStructConcreteContracts(t *testing.T) {
	t.Run("integer", func(t *testing.T) { requireInvalidContract[int](t) })
	t.Run("slice", func(t *testing.T) { requireInvalidContract[[]int](t) })
	t.Run("map", func(t *testing.T) { requireInvalidContract[map[string]int](t) })
	t.Run("function", func(t *testing.T) { requireInvalidContract[func()](t) })
	t.Run("channel", func(t *testing.T) { requireInvalidContract[chan int](t) })
}

func requireContractAcceptance[Carrier, Contract any](t *testing.T, want bool) {
	t.Helper()
	charter, err := sail.CompileLegend[acceptanceLegend[Contract]]()
	if err != nil {
		t.Fatalf("compile contract: %v", err)
	}
	_, err = sail.Populate[acceptanceSubject[Carrier]](charter)
	if want && err != nil {
		t.Fatalf("compatible carrier rejected: %v", err)
	}
	if !want && err == nil {
		t.Fatal("incompatible carrier accepted")
	}
}

func requireInvalidContract[Contract any](t *testing.T) {
	t.Helper()
	if _, err := sail.CompileLegend[acceptanceLegend[Contract]](); err == nil {
		t.Fatal("invalid concrete contract compiled")
	}
}
