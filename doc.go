// Package sail compiles Charters from user-authored Legends, populates Charts
// from Subject declarations, and materializes Vessels holding live Subjects.
// Charter and Vessel traversal entry points accept configurations from
// github.com/qlustra/sail/chart. Vessel.Propagate walks live carriers with
// interceptors from github.com/qlustra/sail/subject. Bind connects populated
// markers to their live carriers: Subject fields, or the Subject root for the
// root Legend.
// Authored declaration vocabulary lives in github.com/qlustra/sail/legend.
package sail
