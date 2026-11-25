// Package sail compiles Go struct models into reusable grammars and applies them
// to other Go struct declarations. A framework authors its grammar on a Legend
// model using markers from [github.com/qlustra/sail/legend]. An application
// declares a Chart using the framework's tags; population produces a fresh
// instance of that same Legend model.
//
// Use [Compile] to obtain a Navigator. Build or Populate a Legend, then use
// Survey to visit its occurrences, Next to select one immediate occurrence, or
// a prepared Course to find complete routes. These operations work on the
// populated model, not a live Chart. A token's [ChartLocation] can be resolved
// separately against a live Chart when a framework needs field access.
package sail
