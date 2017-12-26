package fsdb

import (
	"fmt"
)

// Make sure *ErrNoSuchKey satisifies error interface.
var _ error = new(ErrNoSuchKey)

// ErrNoSuchKey is the error returned by Read and Delete functions when the key
// requested does not exists.
type ErrNoSuchKey struct {
	key Key
}

func (err *ErrNoSuchKey) Error() string {
	return fmt.Sprintf("no such key: %q", err.key)
}

// NewErrNoSuchKey creates a new ErrNoSuchKey error
func NewErrNoSuchKey(key Key) *ErrNoSuchKey {
	return &ErrNoSuchKey{
		key: key,
	}
}

// IsErrNoSuchKey checks whether a given error is ErrNoSuchKey error.
func IsErrNoSuchKey(err error) bool {
	_, ok := err.(*ErrNoSuchKey)
	return ok
}
