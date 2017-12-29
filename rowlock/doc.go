// Package rowlock provides an implementation of row lock.
//
// A row lock is a set of locks associated with rows.
// Instead of locking and unlocking globally,
// you only operate locks on a row level.
package rowlock
