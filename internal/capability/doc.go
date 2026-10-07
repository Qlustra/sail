// Package capability compiles declaration contracts, resolves their carrier
// types, and invokes metadata methods on zero-valued surrogates. Panic recovery
// is confined to declaration hooks; live subject invocation is caller-owned.
package capability
