package legend

import "github.com/qlustra/sail/internal/declaration"

// Self is a capability signature placeholder for the carrier type.
// Quality, sail.ReadCapability, and AcceptContract resolve carrier pointers
// before matching Self, retaining the identity of the defined type. Self may
// appear as a method's parameter or result; compound types containing Self are
// unsupported. It is a contract placeholder, never the stored value type.
type Self = declaration.Self
