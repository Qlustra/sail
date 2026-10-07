package subject

import subjectcore "github.com/qlustra/sail/internal/subject"

// Interceptor registers an ordinary Go interface and its typed callback.
// Construct one with Intercept; a zero Interceptor is invalid.
type Interceptor = subjectcore.Interceptor

// Intercept binds the original live carrier to I before calling intercept.
// I must be an ordinary Go interface; Self templates are not specialized.
// Addressable concrete values prefer their pointer receiver. Dynamic interface
// values retain ordinary Go value semantics; Sail never copies a value to
// manufacture pointer methods. Invalid registrations are reported by Propagate.
func Intercept[I any](intercept func(I, Encounter) (Route, error)) Interceptor {
	return subjectcore.Intercept(intercept)
}
