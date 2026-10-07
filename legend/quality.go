package legend

import "github.com/qlustra/sail/internal/declaration"

// Quality stores a capability result required by its owning qualifier. Declare
// it inside a qualifier composite with sail:"quality" and populate:"Method".
// An absent qualifier leaves its qualities inactive; a present qualifier
// requires every declared quality's capability on the carrier type.
//
// V is the value type and I is an interface, optionally using Self in its
// method signatures. Use any to retain results whose type depends on Self.
// Methods run on fresh zero-valued surrogates during population, not on live
// Subject values. Get and Present only read the stored result.
type Quality[V, I any] = declaration.Quality[V, I]
