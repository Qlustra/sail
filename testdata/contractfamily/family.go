package contractfamily

import "reflect"

// Provider seals the typed witness inside the carrier's owning package.
type Provider interface {
	contract() reflect.Type
}

type API[T any] interface {
	Provider
	Get() T
	Set(T)
}

type Cell[T any] struct {
	WitnessVisits int
}

func (Cell[T]) Get() T { panic("Get must not run during acceptance") }
func (*Cell[T]) Set(T) { panic("Set must not run during acceptance") }

func (c *Cell[T]) contract() reflect.Type {
	var _ API[T] = (*Cell[T])(nil)
	if c.WitnessVisits != 0 {
		panic("witness receiver was not fresh and zero-valued")
	}
	c.WitnessVisits++
	return reflect.TypeFor[API[T]]()
}

type Family struct{}

func (Family) ContractOf(c Provider) reflect.Type { return c.contract() }
