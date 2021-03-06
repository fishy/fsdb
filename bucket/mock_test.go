package bucket

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fishy/fsdb"
)

func TestMock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx := context.Background()

	delay := time.Millisecond * 50
	total := delay * 2
	shorter := time.Millisecond * 30
	longer := time.Millisecond * 60

	key := "foo"
	data := "bar"

	mockDelay := MockOperationDelay{
		Before: delay,
		After:  delay,
	}

	root, err := ioutil.TempDir("", "bucket_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)
	mock := MockBucket(root)
	db := mock.db
	mock.WriteDelay = mockDelay
	mock.DeleteDelay = mockDelay

	// Make sure it's empty
	keys := scanKeys(t, db)
	if len(keys) > 0 {
		t.Fatalf("mock fsdb not empty: %v", keys)
	}

	// Write test
	started := time.Now()
	go func() {
		time.Sleep(shorter)
		keys := scanKeys(t, db)
		if len(keys) > 0 {
			t.Errorf(
				"write operation should delay %v, but already got data after %v: %v",
				delay,
				time.Now().Sub(started),
				keys,
			)
		}
	}()
	go func() {
		time.Sleep(longer)
		keys := scanKeys(t, db)
		if len(keys) != 1 || !keys[0].Equals(fsdb.Key(key)) {
			t.Errorf(
				"write operation should finished after %v, expected key %v, got %v",
				time.Now().Sub(started),
				key,
				keys,
			)
		}
	}()
	if err := mock.Write(ctx, key, strings.NewReader(data)); err != nil {
		t.Errorf("write failed: %v", err)
	}
	elapsed := time.Now().Sub(started)
	if elapsed <= total {
		t.Errorf(
			"write function should return after %v, actual time %v",
			total,
			elapsed,
		)
	}

	// Delete test
	started = time.Now()
	go func() {
		time.Sleep(shorter)
		keys := scanKeys(t, db)
		if len(keys) != 1 || !keys[0].Equals(fsdb.Key(key)) {
			t.Errorf(
				"delete operation should delay %v, but got %v after %v, expected %v",
				delay,
				keys,
				time.Now().Sub(started),
				key,
			)
		}
	}()
	go func() {
		time.Sleep(longer)
		keys := scanKeys(t, db)
		if len(keys) > 0 {
			t.Errorf(
				"delete operation should finished after %v, but got keys %v",
				time.Now().Sub(started),
				keys,
			)
		}
	}()
	if err := mock.Delete(ctx, key); err != nil {
		t.Errorf("delete failed: %v", err)
	}
	elapsed = time.Now().Sub(started)
	if elapsed <= total {
		t.Errorf(
			"delete function should return after %v, actual time %v",
			total,
			elapsed,
		)
	}
}

func TestTotal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx := context.Background()

	total := time.Millisecond * 50
	shorter := time.Millisecond * 30

	key := "foo"
	data := "bar"

	root, err := ioutil.TempDir("", "bucket_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)
	mock := MockBucket(root)
	db := mock.db
	mock.WriteDelay = MockOperationDelay{
		Before: shorter,
		Total:  total,
	}
	mock.ReadDelay = MockOperationDelay{
		Before: shorter,
		After:  shorter,
		Total:  total,
	}

	// Make sure it's empty
	keys := scanKeys(t, db)
	if len(keys) > 0 {
		t.Fatalf("mock fsdb not empty: %v", keys)
	}

	// Write test
	started := time.Now()
	if err := mock.Write(ctx, key, strings.NewReader(data)); err != nil {
		t.Errorf("write failed: %v", err)
	}
	elapsed := time.Now().Sub(started)
	if elapsed <= total {
		t.Errorf(
			"write function should return after %v, actual time %v",
			total,
			elapsed,
		)
	}

	// Delete test
	started = time.Now()
	closer, err := mock.Read(ctx, key)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	defer closer.Close()
	elapsed = time.Now().Sub(started)
	if elapsed <= total {
		t.Errorf(
			"delete function should return after %v, actual time %v",
			total,
			elapsed,
		)
	}
	if elapsed <= shorter*2 {
		t.Errorf(
			"delete function should return after %v, actual time %v",
			shorter*2,
			elapsed,
		)
	}
}

func scanKeys(t *testing.T, db fsdb.Local) []fsdb.Key {
	t.Helper()
	ctx := context.Background()
	keys := make([]fsdb.Key, 0)
	keyFunc := func(key fsdb.Key) bool {
		keys = append(keys, key)
		return true
	}
	if err := db.ScanKeys(ctx, keyFunc, fsdb.IgnoreAll); err != nil {
		t.Fatalf("ScanKeys returned error: %v", err)
	}
	return keys
}
