package sail

import "github.com/qlustra/sail/internal/chart/values"

// ChartLocation identifies a field relative to the exact Chart type used for
// population. Index paths use Go's Chart field positions: ignored fields still
// count, and promoted paths include embedding steps. It holds no live Chart
// reference. Resolve accepts a non-nil pointer to that exact root type and returns
// checked access to its actual field, without traversal or operation dispatch.
type ChartLocation = values.ChartLocation
