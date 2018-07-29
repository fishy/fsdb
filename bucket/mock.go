package bucket

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/fishy/fsdb"
	"github.com/fishy/fsdb/local"
)

// Make sure *Mock satisfies Bucket interface.
var _ Bucket = (*Mock)(nil)

// MockOperationDelay defines the delays of an operation (function call).
// It's useful to mimic network latency in local tests.
type MockOperationDelay struct {
	// Before is the delay between the function call and the actual operation.
	Before time.Duration

	// After is the delay between the actual operation completes and the function
	// returns.
	After time.Duration

	// Total is the minimal time the function call should take before returning.
	// Here are some examples assuming the operation itself takes 1ms and we have
	// a Total set to 50ms:
	// - If both Before and After are set to zero, the function call will take
	//   1ms to do the actual operation, then wait for ~49ms for the 50ms Total to
	//   pass, so it returns after approximately 50ms;
	// - If both Before and After are set to 30ms, the function call will first
	//   sleep 30ms, then take 1ms to do the actual operation, then sleep 30ms
	//   again. At this time the Total 50ms is already passed so the function
	//   returns at approximately 61ms.
	Total time.Duration
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
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		time.Sleep(m.ReadDelay.Total)
	}()

	time.Sleep(m.ReadDelay.Before)
	defer time.Sleep(m.ReadDelay.After)
	return m.db.Read(ctx, fsdb.Key(name))
}

// Write writes the file to fsdb.
func (m *Mock) Write(ctx context.Context, name string, data io.Reader) error {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		time.Sleep(m.WriteDelay.Total)
	}()

	time.Sleep(m.WriteDelay.Before)
	defer time.Sleep(m.WriteDelay.After)
	return m.db.Write(ctx, fsdb.Key(name), data)
}

// Delete deletes the file from fsdb.
func (m *Mock) Delete(ctx context.Context, name string) error {
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		time.Sleep(m.DeleteDelay.Total)
	}()

	time.Sleep(m.DeleteDelay.Before)
	defer time.Sleep(m.DeleteDelay.After)
	return m.db.Delete(ctx, fsdb.Key(name))
}

// IsNotExist calls fsdb.IsNoSuchKeyError.
func (m *Mock) IsNotExist(err error) bool {
	return fsdb.IsNoSuchKeyError(err)
}
