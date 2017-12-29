package bucket

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fishy/fsdb/interface"
)

func TestMock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

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
	if err := mock.Write(key, strings.NewReader(data)); err != nil {
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
	if err := mock.Delete(key); err != nil {
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

func scanKeys(t *testing.T, db fsdb.Local) []fsdb.Key {
	t.Helper()
	keys := make([]fsdb.Key, 0)
	keyFunc := func(key fsdb.Key) bool {
		keys = append(keys, key)
		return true
	}
	if err := db.ScanKeys(keyFunc, fsdb.IgnoreAll); err != nil {
		t.Fatalf("ScanKeys returned error: %v", err)
	}
	return keys
}
