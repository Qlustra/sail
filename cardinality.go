package sail

import "fmt"

type cardinality struct {
	min int
	max int // -1 means unbounded.
}

func compileCardinality(raw string) (cardinality, error) {
	switch raw {
	case "any":
		return cardinality{min: 0, max: -1}, nil
	case "exactly-one":
		return cardinality{min: 1, max: 1}, nil
	case "at-least-one":
		return cardinality{min: 1, max: -1}, nil
	case "at-most-one":
		return cardinality{min: 0, max: 1}, nil
	default:
		return cardinality{}, fmt.Errorf("invalid or missing cardinality %q", raw)
	}
}
