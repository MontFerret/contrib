// Package cssx compiles and validates CSSX expressions for HTML drivers.
//
// CSSX pseudo-functions operate on normalized selections. Maps preserve one slot
// per input, traversals flat-map nodes without implicit deduplication, filters
// retain matching nodes, and selection operators keep selection shape. Only
// reducers and cardinality operations collapse a selection.
package cssx
