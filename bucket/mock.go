package bucket

import (
	"context"
	"io"
	"time"

	"github.com/fishy/fsdb"
	"github.com/fishy/fsdb/local"
)

// Make sure *Mock satisfies Bucket interface.
var _ Bucket = (*Mock)(nil)

// MockOperationDelay defines the delay before and after an operation.
// It's useful to mimic network latency in tests.
type MockOperationDelay struct {
	// Before is the delay between the function call and the actual operation.
	Before time.Duration

	// After is the delay between the actual operation completes and the function
	// returns.
	After time.Duration
}

// Mock is a mock implementation of Bucket, backed by local FSDB.
type Mock struct {
	db fsdb.Local

	ReadDelay   MockOperationDelay
	WriteDelay  MockOperationDelay
	DeleteDelay MockOperationDelay
}

// MockBucket creates a new mock Bucket using fsdb.
func MockBucket(root string) *Mock {
	return &Mock{
		db: local.Open(local.NewDefaultOptions(root)),
	}
}

// Read reads the file from fsdb.
func (m *Mock) Read(ctx context.Context, name string) (io.ReadCloser, error) {
	time.Sleep(m.ReadDelay.Before)
	defer time.Sleep(m.ReadDelay.After)
	return m.db.Read(ctx, fsdb.Key(name))
}

// Write writes the file to fsdb.
func (m *Mock) Write(ctx context.Context, name string, data io.Reader) error {
	time.Sleep(m.WriteDelay.Before)
	defer time.Sleep(m.WriteDelay.After)
	return m.db.Write(ctx, fsdb.Key(name), data)
}

// Delete deletes the file from fsdb.
func (m *Mock) Delete(ctx context.Context, name string) error {
	time.Sleep(m.DeleteDelay.Before)
	defer time.Sleep(m.DeleteDelay.After)
	return m.db.Delete(ctx, fsdb.Key(name))
}

// IsNotExist calls fsdb.IsNoSuchKeyError.
func (m *Mock) IsNotExist(err error) bool {
	return fsdb.IsNoSuchKeyError(err)
}
