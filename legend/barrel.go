package legend

import "github.com/qlustra/sail/internal/declaration"

// Barrel stores the result of a population-time capability method. V is the
// destination type and I is the capability interface. Compilation validates
// these types and the method named by the field's populate tag.
//
// A barrel is optional unless its field declares barrel:"required". Get and
// Present only read stored state; they never invoke or validate the capability.
// Each populated barrel invokes its method on a fresh zero-valued surrogate;
// the method must not depend on initialized Subject state.
type Barrel[V, I any] = declaration.Barrel[V, I]
