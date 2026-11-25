package sail

import (
	"strings"
	"testing"
)

func TestPlotPrepareBoundary(t *testing.T) {
	chart, err := Compile[idLegend]()
	if err != nil {
		t.Fatal(err)
	}
	plot := chart.PlotCourse()
	if course, err := plot.Prepare(); err == nil || course != nil {
		t.Fatal("empty plot published a Course")
	}
	plot.Waypoint(Select[any]())
	course, err := plot.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if course.navigator != chart || len(course.steps) != 1 {
		t.Fatal("Course must retain its Navigator and an independent waypoint sequence")
	}
	plot.Waypoint(Select[any]().InLane(""))
	if invalid, err := plot.Prepare(); err == nil || invalid != nil || !strings.Contains(err.Error(), "waypoint 2") {
		t.Fatalf("invalid appended selector: %v, %v", invalid, err)
	}
	if len(course.steps) != 1 {
		t.Fatal("editing the Plot changed the prepared Course")
	}
}
