package main

import (
	"fmt"
	"io"
	"os"

	"github.com/qlustra/sail"
	"github.com/qlustra/sail/legend"
)

type PipelineLegend struct {
	legend.Legend `sail:"legend" mark:"pipeline"`
	Jobs          JobLane `sail:"lane"`
}

type JobLane struct {
	legend.Lane
	Items []Job `sail:"spot" spot:"multi,required"`
}

type Job struct {
	legend.Landmark `sail:"landmark" family:"job"`
	Name            legend.Feature `sail:"feature" symbol:"name"`
}

type ReleaseChart struct {
	Compile struct{} `pipeline:"job" name:"compile"`
}

func run(w io.Writer) error {
	navigator, err := sail.Compile[PipelineLegend]()
	if err != nil {
		return err
	}
	model, err := navigator.Build[ReleaseChart]()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, model.Jobs.Items[0].Name.Get())
	return err
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
