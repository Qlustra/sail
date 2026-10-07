package legend

import "github.com/qlustra/sail/internal/declaration"

// Self is a capability signature placeholder for the receiving Subject type.
// Quality, sail.ReadCapability, and AcceptContract resolve Subject pointers
// before matching Self, retaining the identity of the defined type. Self may
// appear as a method's parameter or result; compound types containing Self are
// unsupported. It is a contract placeholder, never the stored value type.
type Self = declaration.Self
