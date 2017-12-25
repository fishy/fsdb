package fsdb

import (
	"fmt"
)

// ErrNoSuchKey is the error returned by Read and Delete functions when the key
// requested does not exists.
type ErrNoSuchKey struct {
	key Key
}

func (err *ErrNoSuchKey) Error() string {
	return fmt.Sprintf("no such key: %q", err.key)
}

// NewErrNoSuchKey creates a new ErrNoSuchKey error
func NewErrNoSuchKey(key Key) error {
	return &ErrNoSuchKey{
		key: key,
	}
}

// IsErrNoSuchKey checks whether a given error is ErrNoSuchKey error.
func IsErrNoSuchKey(err error) bool {
	_, ok := err.(*ErrNoSuchKey)
	return ok
}
