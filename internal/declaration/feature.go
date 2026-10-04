package declaration

type Feature struct {
	value string
}

func (Feature) feature() {}

type FeatureDeclaration interface {
	feature()
}

func (f *Feature) Get() string { return f.value }

func (f *Feature) Set(value string) { f.value = value }
