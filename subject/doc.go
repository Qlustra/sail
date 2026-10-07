// Package subject provides propagation through live Sail Subjects. Typed
// interceptors receive existing carrier receivers and return Routes controlling
// local visit order, descent, pruning, and stopping. Use an explicit Order and a
// Propagation interceptor list with sail.Vessel.Propagate.
//
// Descend may supply successful completion work. Enterable carriers add typed
// members through Iterate; those members join the same walk. Encounters expose
// reliable Identity, First-arrival history, and Cycle ancestry. Cyclic arrivals
// can be handled, but Sail rejects descent through a cycle. Errors stop at the
// first failure; Stop, errors, and panics cancel pending completions.
package subject
