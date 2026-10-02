package sail

import "reflect"

// Barrel stores the result of a population-time capability method. V is the
// destination type and I is the capability interface. Exploration validates
// these types and the method named by the field's populate tag.
//
// A barrel is optional unless its field declares barrel:"required". Get and
// Present only read stored state; they never invoke or validate the capability.
// Each populated barrel invokes its method on a fresh zero-valued surrogate;
// the method must not depend on initialized application state.
type Barrel[V, I any] struct {
	value   V
	present bool
}

func (b Barrel[V, I]) Get() V { return b.value }

// Present distinguishes an unavailable capability from a populated zero value.
func (b Barrel[V, I]) Present() bool { return b.present }

type barrelShape struct {
	value      reflect.Type
	capability reflect.Type
	carrier    reflect.Type
}

// The sealed shape keeps both generic arguments available to exploration.
// It does not dereference the receiver, including an embedded nil *Barrel.
func (*Barrel[V, I]) barrel() barrelShape {
	return barrelShape{
		value:      reflect.TypeFor[V](),
		capability: reflect.TypeFor[I](),
		carrier:    reflect.TypeFor[Barrel[V, I]](),
	}
}

func (b *Barrel[V, I]) fillBarrel(value reflect.Value) {
	// Set preserves a nil interface result and applies the assignability
	// contract checked during exploration, without a runtime type assertion.
	reflect.ValueOf(&b.value).Elem().Set(value)
	b.present = true
}

type barrelDeclaration interface {
	barrel() barrelShape
	fillBarrel(reflect.Value)
}
