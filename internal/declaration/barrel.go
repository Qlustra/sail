package declaration

import "reflect"

type Barrel[V, I any] struct {
	value   V
	present bool
}

func (b Barrel[V, I]) Get() V { return b.value }

// Present distinguishes an unavailable capability from a populated zero value.
func (b Barrel[V, I]) Present() bool { return b.present }

type BarrelShape struct {
	Value      reflect.Type
	Capability reflect.Type
	Base       reflect.Type // Concrete Barrel type embedded by composite declarations.
}

// The sealed shape keeps both generic arguments available to compilation.
// It does not dereference the receiver, including an embedded nil *Barrel.
func (*Barrel[V, I]) barrel() BarrelShape {
	return BarrelShape{
		Value:      reflect.TypeFor[V](),
		Capability: reflect.TypeFor[I](),
		Base:       reflect.TypeFor[Barrel[V, I]](),
	}
}

func (b *Barrel[V, I]) fillBarrel(value reflect.Value) {
	// Set preserves nil interface results and uses the checked assignability
	// contract without a potentially failing type assertion.
	reflect.ValueOf(&b.value).Elem().Set(value)
	b.present = true
}

type BarrelDeclaration interface {
	barrel() BarrelShape
	fillBarrel(reflect.Value)
}

func DescribeBarrel(provider BarrelDeclaration) BarrelShape { return provider.barrel() }

func FillBarrel(destination BarrelDeclaration, value reflect.Value) {
	destination.fillBarrel(value)
}
