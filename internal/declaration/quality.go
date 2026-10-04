package declaration

import "reflect"

type Quality[V, I any] struct {
	value   V
	present bool
}

func (q Quality[V, I]) Get() V { return q.value }

// Present distinguishes an inactive quality from a successfully populated
// zero or nil value.
func (q Quality[V, I]) Present() bool { return q.present }

type QualityShape struct {
	Value      reflect.Type
	Capability reflect.Type
	Carrier    reflect.Type
}

func (*Quality[V, I]) quality() QualityShape {
	return QualityShape{
		Value:      reflect.TypeFor[V](),
		Capability: reflect.TypeFor[I](),
		Carrier:    reflect.TypeFor[Quality[V, I]](),
	}
}

func (q *Quality[V, I]) fillQuality(value reflect.Value) {
	reflect.ValueOf(&q.value).Elem().Set(value)
	q.present = true
}

type QualityDeclaration interface {
	quality() QualityShape
	fillQuality(reflect.Value)
}

func DescribeQuality(provider QualityDeclaration) QualityShape { return provider.quality() }

func FillQuality(destination QualityDeclaration, value reflect.Value) {
	destination.fillQuality(value)
}
