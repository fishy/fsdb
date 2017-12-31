package rowlock

import (
	"sync"

	"github.com/fishy/fsdb/libs/pool"
)

// LockerPoolMaxSize is the max size of the locker pool.
//
// The locker pool contains the lockers to be used for new rows.
// It has no relation to the number of rows in the RowLock.
const LockerPoolMaxSize = 10

// NewLocker defines a type of function that can be used to create a new Locker.
type NewLocker func() sync.Locker

// MutexNewLocker is a NewLocker using sync.Mutex.
func MutexNewLocker() sync.Locker {
	return new(sync.Mutex)
}

// RowLock defines a set of row lock.
//
// A set of row lock is a set of locks.
// When you do Lock/Unlock operations, you don't do them on a glogal scale.
// Instead, a Lock/Unlock operation is operated on a given row/key.
type RowLock struct {
	locks      sync.Map
	g          pool.Generator
	lockerPool *pool.Pool
}

// NewRowLock creates a new RowLock with the given NewLocker.
func NewRowLock(f NewLocker) *RowLock {
	return &RowLock{
		g: func() interface{} {
			return f()
		},
		lockerPool: pool.NewPool(LockerPoolMaxSize),
	}
}

// Lock locks a row.
//
// row must be hashable.
func (rl *RowLock) Lock(row interface{}) {
	rl.getLocker(row).Lock()
}

// Unlock unlocks a row.
//
// row must be hashable.
func (rl *RowLock) Unlock(row interface{}) {
	rl.getLocker(row).Unlock()
}

// getLocker returns the lock for the given row.
//
// If this is a new row,
// a new locker will be created using the NewLocker specified in NewRowLock.
func (rl *RowLock) getLocker(row interface{}) sync.Locker {
	newLocker := rl.lockerPool.Get(rl.g)
	locker, loaded := rl.locks.LoadOrStore(row, newLocker)
	if loaded {
		rl.lockerPool.Put(newLocker)
	}
	return locker.(sync.Locker)
}
