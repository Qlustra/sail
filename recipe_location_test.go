package sail_test

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/qlustra/intel/capture"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type locationFlag struct {
	legend.Landmark `sail:"landmark" family:"flag"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}
type locationCommand struct {
	legend.Region `sail:"region" family:"command"`
}
type locationLane struct {
	legend.Lane
	Flag    locationFlag    `sail:"spot" spot:"single"`
	Command locationCommand `sail:"spot" spot:"single"`
}
type locationLegend struct {
	legend.Legend `sail:"legend" mark:"loc"`
	Items         locationLane `sail:"lane"`
}
type locationHandler struct{ Count int }

func (h *locationHandler) Run(delta int) int { h.Count += delta; return h.Count }

type locationRecipe struct {
	Ignored     string
	Destination string          `loc:"flag" name:"output"`
	Handler     locationHandler `loc:"command"`
	Other       string
}
type locationOtherRecipe locationRecipe
type locationAlias = locationRecipe

func locationFor(t *testing.T, item occurrenceReader) *sail.ChartLocation {
	t.Helper()
	location, check := item.ChartLocation()
	if !check.OK() || location == nil {
		t.Fatalf("location: %v", check)
	}
	return location
}

func locationAddress(t *testing.T, location *sail.ChartLocation, chart any) any {
	t.Helper()
	field, check := location.Resolve(chart)
	if !check.OK() || field == nil {
		t.Fatalf("resolve: %v", check)
	}
	address, check := field.Address()
	if !check.OK() {
		t.Fatalf("address: %v", check)
	}
	return address
}

func TestChartLocationLiveFieldAndType(t *testing.T) {
	chart, err := sail.Compile[locationLegend]()
	if err != nil {
		t.Fatal(err)
	}
	app := locationRecipe{Ignored: "untouched", Destination: "configured", Other: "neighbor"}
	previousLegend, err := chart.Populate(app)
	if err != nil {
		t.Fatal(err)
	}
	location := locationFor(t, previousLegend.Items.Flag)
	if previousLegend.Items.Flag.Name.Get() != "output" || !reflect.DeepEqual(location.Index(), []int{1}) || location.ChartType().CapturedType() != reflect.TypeFor[locationRecipe]() {
		t.Fatal("incorrect flag metadata")
	}
	address, ok := locationAddress(t, location, &app).(*string)
	if !ok || address != &app.Destination || *address != "configured" {
		t.Fatal("did not resolve the configured field")
	}
	*address = "changed"
	second := locationRecipe{Destination: "second", Other: "second neighbor"}
	secondAddress, ok := locationAddress(t, location, &second).(*string)
	if !ok || secondAddress != &second.Destination || *secondAddress != "second" {
		t.Fatal("second field not live")
	}
	*secondAddress = "changed second"
	if app.Destination != "changed" || app.Ignored != "untouched" || app.Other != "neighbor" || second.Destination != "changed second" || second.Other != "second neighbor" {
		t.Fatal("wrong field or instance mutated")
	}
	if previousLegend.Items.Flag.Name.Get() != "output" || !reflect.DeepEqual(locationFor(t, previousLegend.Items.Flag).Index(), []int{1}) {
		t.Fatal("resolution changed previous Legend metadata")
	}

	other := locationOtherRecipe{Destination: "other", Other: "safe"}
	var nilRecipe *locationRecipe
	pointer := &app
	for _, target := range []any{nil, nilRecipe, app, &pointer, &other} {
		field, check := location.Resolve(target)
		if check.OK() || field != nil || !strings.Contains(check.Why(), "root type") {
			t.Fatalf("invalid root %T: field=%v check=%v", target, field, check)
		}
	}
	if other.Destination != "other" || other.Other != "safe" {
		t.Fatal("rejected root mutated")
	}
	if location.ChartType().CapturedType() != reflect.TypeFor[locationRecipe]() || !reflect.DeepEqual(location.Index(), []int{1}) || !reflect.DeepEqual(locationFor(t, previousLegend.Items.Flag).Index(), []int{1}) {
		t.Fatal("rejected roots changed location metadata")
	}
	otherLegend, err := chart.Build[locationOtherRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	otherAddress, ok := locationAddress(t, locationFor(t, otherLegend.Items.Flag), &other).(*string)
	if !ok || otherAddress != &other.Destination || *otherAddress != "other" {
		t.Fatal("defined root did not resolve against its own location")
	}
	alias := locationAlias{Destination: "alias"}
	aliasAddress, ok := locationAddress(t, location, &alias).(*string)
	if !ok || aliasAddress != &alias.Destination || *aliasAddress != "alias" {
		t.Fatal("alias not accepted")
	}
}

func TestChartLocationNestedPromotion(t *testing.T) {
	chart, err := sail.Compile[occurrenceNestedLegend]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[occurrenceNestedRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	if len(populatedLegend.Roots.Regions) != 2 || len(populatedLegend.Roots.Regions[0].Children.Leaves) != 2 {
		t.Fatal("missing nested occurrences")
	}
	location := locationFor(t, populatedLegend.Roots.Regions[0].Children.Leaves[0])
	if !reflect.DeepEqual(location.Index(), []int{1, 1, 1, 1}) || location.ChartType().CapturedType() != reflect.TypeFor[occurrenceNestedRecipe]() {
		t.Fatalf("wrong root/path: %v", location.Index())
	}
	app := occurrenceNestedRecipe{}
	app.First.Leaf = true
	app.Second.Leaf = true
	address, ok := locationAddress(t, location, &app).(*bool)
	if !ok || address != &app.First.Leaf || !*address {
		t.Fatal("nested promoted field not resolved")
	}
	*address = false
	if app.First.Leaf || !app.Second.Leaf {
		t.Fatal("wrong nested field changed")
	}
}

func TestChartLocationFrameworkInvocation(t *testing.T) {
	chart, err := sail.Compile[locationLegend]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[locationRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	app := locationRecipe{Handler: locationHandler{Count: 12}}
	location := locationFor(t, populatedLegend.Items.Command)
	if !reflect.DeepEqual(location.Index(), []int{2}) {
		t.Fatalf("handler path: %v", location.Index())
	}
	field, check := location.Resolve(&app)
	if !check.OK() || field == nil {
		t.Fatal(check)
	}
	address, check := field.Address()
	if !check.OK() || address != &app.Handler {
		t.Fatal("not the live receiver", check)
	}
	receiver, check := field.AsStruct()
	if !check.OK() || receiver == nil {
		t.Fatal(check)
	}
	// Framework test-owned operation: Sail only resolves the location.
	result, check := receiver.Invoke[interface{ Run(int) int }, int]("Run", func(h interface{ Run(int) int }) (int, logic.Check) { return h.Run(3), logic.Ok() })
	if !check.OK() || result != 15 || app.Handler.Count != 15 {
		t.Fatalf("live invocation: result=%d count=%d check=%v", result, app.Handler.Count, check)
	}
}

func TestChartLocationUnavailable(t *testing.T) {
	var zero sail.ChartLocation
	var nilLocation *sail.ChartLocation
	for _, location := range []*sail.ChartLocation{&zero, nilLocation} {
		field, check := location.Resolve(&locationRecipe{})
		if check.OK() || field != nil || !strings.Contains(check.Why(), "location") {
			t.Fatalf("unavailable location: %v", check)
		}
	}
	chart, err := sail.Compile[locationLegend]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[occurrenceAbsentRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	location, check := populatedLegend.Items.Flag.ChartLocation()
	if populatedLegend.Items.Flag.Present() || check.OK() || location != nil {
		t.Fatalf("optional flag has a location: %v", check)
	}
}

func TestChartLocationConcurrentIndependentCharts(t *testing.T) {
	chart, err := sail.Compile[locationLegend]()
	if err != nil {
		t.Fatal(err)
	}
	populatedLegend, err := chart.Build[locationRecipe]()
	if err != nil {
		t.Fatal(err)
	}
	rootLocation := locationFor(t, populatedLegend.Items.Flag)
	const n = 24
	apps := make([]locationRecipe, n)
	var wg sync.WaitGroup
	errors := make(chan string, n)
	for i := range apps {
		apps[i].Destination = "initial"
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			field, check := rootLocation.Resolve(&apps[i])
			if !check.OK() || field == nil {
				errors <- "resolve failed"
				return
			}
			address, check := field.Address()
			if !check.OK() {
				errors <- "address failed"
				return
			}
			ptr, ok := address.(*string)
			if !ok || ptr != &apps[i].Destination || *ptr != "initial" {
				errors <- "wrong live field"
				return
			}
			*ptr = "updated"
		}(i)
	}
	wg.Wait()
	close(errors)
	for message := range errors {
		t.Error(message)
	}
	for i := range apps {
		if apps[i].Destination != "updated" {
			t.Errorf("app %d not updated", i)
		}
	}
	if rootLocation.ChartType().CapturedType() != reflect.TypeFor[locationRecipe]() || !reflect.DeepEqual(rootLocation.Index(), []int{1}) || !reflect.DeepEqual(locationFor(t, populatedLegend.Items.Flag).Index(), []int{1}) {
		t.Fatal("location metadata changed")
	}
}

// Assert the public resolver returns Intel's existing checked handle.
var _ func(*sail.ChartLocation, any) (*capture.ValueInterrogator, logic.Check) = (*sail.ChartLocation).Resolve
