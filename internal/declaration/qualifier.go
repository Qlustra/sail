package declaration

type Qualifier struct {
	present bool
}

func (Qualifier) qualifier() {}

type QualifierDeclaration interface {
	qualifier()
}

func (q *Qualifier) Present() bool { return q.present }

func (q *Qualifier) Set() { q.present = true }

func (q *Qualifier) Unset() { q.present = false }
