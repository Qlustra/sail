package legend

import "github.com/qlustra/sail/internal/declaration"

// Qualifier records whether its symbol is present among the mark's qualifiers.
type Qualifier = declaration.Qualifier

type MarkQualifier struct {
	Symbol string
}

type QualifierDeclaration interface {
	Symbol() string
}
