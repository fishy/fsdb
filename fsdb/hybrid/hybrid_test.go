package hybrid_test

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fishy/fsdb/bucket"
	"github.com/fishy/fsdb/fsdb"
	"github.com/fishy/fsdb/fsdb/hybrid"
	"github.com/fishy/fsdb/fsdb/local"
)

type dbCollection struct {
	DB     fsdb.FSDB
	Local  fsdb.Local
	Remote *bucket.Mock
	Opts   hybrid.OptionsBuilder
}

func (db *dbCollection) Open(ctx context.Context) {
	db.DB = hybrid.Open(ctx, db.Local, db.Remote, db.Opts)
}

func TestLocal(t *testing.T) {
	root, db := createHybridDB(t, "local: ")
	defer os.RemoveAll(root)
	ctx := context.Background()
	db.Open(ctx)

	key := fsdb.Key("foo")
	content := "bar"

	if _, err := db.DB.Read(ctx, key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"read from empty hybrid db should return NoSuchKeyError, got %v",
			err,
		)
	}

	if err := db.DB.Write(ctx, key, strings.NewReader(content)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	compareContent(t, db.DB, key, content)

	if err := db.DB.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := db.DB.Read(ctx, key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"read from empty hybrid db should return NoSuchKeyError, got %v",
			err,
		)
	}
}

func TestHybrid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	delay := time.Millisecond * 100
	longer := time.Millisecond * 150

	root, db := createHybridDB(t, "hybrid: ")
	defer os.RemoveAll(root)
	db.Opts.SetUploadDelay(delay).SetSkipFunc(hybrid.UploadAll)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db.Open(ctx)

	key := fsdb.Key("foo")
	content := "bar"

	if _, err := db.DB.Read(ctx, key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"read from empty hybrid db should return NoSuchKeyError, got %v",
			err,
		)
	}

	if err := db.DB.Write(ctx, key, strings.NewReader(content)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	time.Sleep(longer)

	if _, err := db.Local.Read(ctx, key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"key should be uploaded to remote and deleted locally, got %v",
			err,
		)
	}

	compareContent(t, db.DB, key, content)
	// Now it should be available locally
	compareContent(t, db.Local, key, content)

	time.Sleep(longer)

	if _, err := db.Local.Read(ctx, key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"key should be uploaded to remote and deleted locally again, got %v",
			err,
		)
	}

	compareContent(t, db.DB, key, content)
	// Now it should be available locally
	compareContent(t, db.Local, key, content)

	if err := db.DB.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := db.DB.Read(ctx, key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"read from empty hybrid db should return NoSuchKeyError, got %v",
			err,
		)
	}
}

func TestSkip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	delay := time.Millisecond * 100
	longer := delay * 2

	key1 := fsdb.Key("foo")
	key2 := fsdb.Key("bar")
	content := "foobar"

	skipFunc := func(key fsdb.Key) bool {
		return key.Equals(key2)
	}

	root, db := createHybridDB(t, "skip: ")
	defer os.RemoveAll(root)
	db.Opts.SetUploadDelay(delay).SetSkipFunc(skipFunc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db.Open(ctx)

	if err := db.DB.Write(ctx, key1, strings.NewReader(content)); err != nil {
		t.Fatalf("Write %v failed: %v", key1, err)
	}
	if err := db.DB.Write(ctx, key2, strings.NewReader(content)); err != nil {
		t.Fatalf("Write %v failed: %v", key2, err)
	}

	time.Sleep(longer)

	if _, err := db.Local.Read(ctx, key1); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf(
			"%v should be uploaded to remote and deleted locally, got %v",
			key1,
			err,
		)
	}
	compareContent(t, db.Local, key2, content)

	compareContent(t, db.DB, key1, content)
	compareContent(t, db.DB, key2, content)
}

func TestSlowUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Write 6 keys, provide 4 threads to upload. After one upload cycle there
	// should be 2 keys left locally.

	delay := time.Millisecond * 100
	// longer should be slightly larger than 2 * delay,
	// as we need one delay before uploading and another delay for uploading.
	longer := time.Millisecond * 250

	keys := []fsdb.Key{
		fsdb.Key("key0"),
		fsdb.Key("key1"),
		fsdb.Key("key2"),
		fsdb.Key("key3"),
		fsdb.Key("key4"),
		fsdb.Key("key5"),
	}
	content := "foobar"
	left := 2

	root, db := createHybridDB(t, "slow-upload: ")
	defer os.RemoveAll(root)
	db.Remote.WriteDelay = bucket.MockOperationDelay{
		Before: delay,
		After:  0,
	}
	db.Opts.SetUploadDelay(delay)
	db.Opts.SetUploadThreadNum(len(keys) - left)
	db.Opts.SetSkipFunc(hybrid.UploadAll)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db.Open(ctx)

	for _, key := range keys {
		if err := db.DB.Write(ctx, key, strings.NewReader(content)); err != nil {
			t.Fatalf("Write %v failed: %v", key, err)
		}
	}

	time.Sleep(longer)
	localKeys := scanKeys(t, db.Local)
	if len(localKeys) != left {
		t.Errorf("Expected %d local keys left, got %v", left, localKeys)
	}
}

func TestUploadRaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Write content1, overwrite with content2 during upload.
	// Check read result after upload finishes.

	delay := time.Millisecond * 100
	// secondWrite should be between delay and 2 * delay
	secondWrite := time.Millisecond * 150
	// readTime should be slightly larger than 2 * delay to make sure the upload
	// finished.
	readTime := time.Millisecond * 250

	key := fsdb.Key("key")
	content1 := "foo"
	content2 := "bar"

	root, db := createHybridDB(t, "upload-race-condition: ")
	defer os.RemoveAll(root)
	db.Remote.WriteDelay = bucket.MockOperationDelay{
		Before: delay,
		After:  0,
	}
	db.Opts.SetUploadDelay(delay).SetSkipFunc(hybrid.UploadAll)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db.Open(ctx)

	if err := db.DB.Write(ctx, key, strings.NewReader(content1)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	go func() {
		time.Sleep(secondWrite)
		if err := db.DB.Write(ctx, key, strings.NewReader(content2)); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		compareContent(t, db.DB, key, content2)
	}()

	time.Sleep(readTime)
	compareContent(t, db.Local, key, content2)
	compareContent(t, db.DB, key, content2)
}

func TestRemoteReadRaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Write content1, wait for upload.
	// Overwrite with content2 during slow read. Check read result.

	delay := time.Millisecond * 100
	longer := time.Millisecond * 120
	secondWrite := 2 * delay

	key := fsdb.Key("key")
	content1 := "foo"
	content2 := "bar"

	root, db := createHybridDB(t, "read-race-condition: ")
	defer os.RemoveAll(root)
	db.Remote.ReadDelay = bucket.MockOperationDelay{
		Before: delay,
		After:  0,
	}
	db.Opts.SetUploadDelay(delay).SetSkipFunc(hybrid.UploadAll)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db.Open(ctx)

	if err := db.DB.Write(ctx, key, strings.NewReader(content1)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	go func() {
		time.Sleep(secondWrite)
		if err := db.DB.Write(ctx, key, strings.NewReader(content2)); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}()

	time.Sleep(longer)
	// When this read finishes, second write already happened
	compareContent(t, db.DB, key, content2)
}

func createHybridDB(
	t *testing.T, prefix string,
) (
	root string, db dbCollection,
) {
	root, err := ioutil.TempDir("", "fsdb_hybrid_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	if !strings.HasSuffix(root, local.PathSeparator) {
		root += local.PathSeparator
	}
	localRoot := root + "local"
	remoteRoot := root + "remote"
	db.Local = local.Open(local.NewDefaultOptions(localRoot))
	db.Remote = bucket.MockBucket(remoteRoot)
	db.Opts = hybrid.NewDefaultOptions()
	db.Opts.SetLogger(log.New(os.Stderr, prefix, log.LstdFlags|log.Lmicroseconds))
	db.Opts.SetSkipFunc(hybrid.SkipAll)
	return
}

func compareContent(t *testing.T, db fsdb.FSDB, key fsdb.Key, content string) {
	t.Helper()

	ctx := context.Background()

	reader, err := db.Read(ctx, key)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	defer reader.Close()
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("read content failed: %v", err)
	}
	if content != string(buf) {
		t.Errorf("read content failed, expected %q, got %q", content, buf)
	}
}

func scanKeys(t *testing.T, db fsdb.Local) []fsdb.Key {
	t.Helper()

	keys := make([]fsdb.Key, 0)
	if err := db.ScanKeys(
		context.Background(),
		func(key fsdb.Key) bool {
			keys = append(keys, key)
			return true
		},
		fsdb.IgnoreAll,
	); err != nil {
		t.Fatalf("ScanKeys returned error: %v", err)
	}
	return keys
}
