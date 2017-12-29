// Package pool provides an implementation of resource pool.
//
// It's implemented as a linked array.
// The Get function does not block on empty pool,
// and the Put function does not block on full pool.
package pool
