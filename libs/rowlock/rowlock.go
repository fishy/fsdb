package rowlock

import (
	"sync"
)

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
	lockerPool sync.Pool
}

// NewRowLock creates a new RowLock with the given NewLocker.
func NewRowLock(f NewLocker) *RowLock {
	return &RowLock{
		lockerPool: sync.Pool{
			New: func() interface{} {
				return f()
			},
		},
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
	newLocker := rl.lockerPool.Get()
	locker, loaded := rl.locks.LoadOrStore(row, newLocker)
	if loaded {
		rl.lockerPool.Put(newLocker)
	}
	return locker.(sync.Locker)
}
