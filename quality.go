package sail

import "reflect"

// Quality stores a capability result required by its owning qualifier. Declare
// it inside a qualifier composite with sail:"quality" and populate:"Method".
// An absent qualifier leaves its qualities inactive; a present qualifier
// requires every declared quality's capability on the receiving application type.
//
// V is the storage type and I is an interface, optionally using Self in its
// method signatures. Use any to retain results whose type depends on Self.
// Methods run on fresh zero-valued surrogates during population, not on live
// application values. Get and Present only read the stored result.
type Quality[V, I any] struct {
	value   V
	present bool
}

func (q Quality[V, I]) Get() V { return q.value }

// Present distinguishes an inactive quality from a successfully populated
// zero or nil value.
func (q Quality[V, I]) Present() bool { return q.present }

type qualityShape struct {
	value      reflect.Type
	capability reflect.Type
	carrier    reflect.Type
}

func (*Quality[V, I]) quality() qualityShape {
	return qualityShape{
		value:      reflect.TypeFor[V](),
		capability: reflect.TypeFor[I](),
		carrier:    reflect.TypeFor[Quality[V, I]](),
	}
}

func (q *Quality[V, I]) fillQuality(value reflect.Value) {
	reflect.ValueOf(&q.value).Elem().Set(value)
	q.present = true
}

type qualityDeclaration interface {
	quality() qualityShape
	fillQuality(reflect.Value)
}
