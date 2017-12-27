package fsdb

import (
	"fmt"
)

// Make sure *NoSuchKeyError satisifies error interface.
var _ error = new(NoSuchKeyError)

// NoSuchKeyError is an error returned by Read and Delete functions when the key
// requested does not exists.
type NoSuchKeyError struct {
	Key Key
}

func (err *NoSuchKeyError) Error() string {
	return fmt.Sprintf("no such key: %q", err.Key)
}

// IsNoSuchKeyError checks whether a given error is NoSuchKeyError.
func IsNoSuchKeyError(err error) bool {
	_, ok := err.(*NoSuchKeyError)
	return ok
}
