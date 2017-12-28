package local_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fishy/fsdb/interface"
	"github.com/fishy/fsdb/local"
)

const lorem = `Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

Ut enim ad minim veniam,
quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

Excepteur sint occaecat cupidatat non proident,
sunt in culpa qui officia deserunt mollit anim id est laborum.`

func TestReadWriteDelete(t *testing.T) {
	root, err := ioutil.TempDir("", "fsdb_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)
	opts := local.NewDefaultOptions(root).SetUseGzip(false)
	db := local.Open(opts)

	key := fsdb.Key("foo")
	// Empty
	testDeleteEmpty(t, db, key)
	testReadEmpty(t, db, key)
	// Write
	testWrite(t, db, key, lorem)
	testRead(t, db, key, lorem)
	testRead(t, db, key, lorem)
	// Overwrite
	content := ""
	testWrite(t, db, key, content)
	testRead(t, db, key, content)
	// Delete
	testDelete(t, db, key)
	testReadEmpty(t, db, key)
}

func TestGzip(t *testing.T) {
	root, err := ioutil.TempDir("", "fsdb_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)
	opts := local.NewDefaultOptions(root).SetUseGzip(true)
	db := local.Open(opts)

	key := fsdb.Key("foo")
	// Empty
	testDeleteEmpty(t, db, key)
	testReadEmpty(t, db, key)
	// Write
	testWrite(t, db, key, lorem)
	testRead(t, db, key, lorem)
	testRead(t, db, key, lorem)
	// Overwrite
	content := ""
	testWrite(t, db, key, content)
	testRead(t, db, key, content)
	// Delete
	testDelete(t, db, key)
	testReadEmpty(t, db, key)
}

func TestChangeCompression(t *testing.T) {
	root, err := ioutil.TempDir("", "fsdb_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)
	gzipOpts := local.NewDefaultOptions(root).SetUseGzip(true)
	gzipDb := local.Open(gzipOpts)

	key := fsdb.Key("foo")
	testWrite(t, gzipDb, key, lorem)
	testRead(t, gzipDb, key, lorem)

	opts := local.NewDefaultOptions(root).SetUseGzip(false)
	db := local.Open(opts)
	testRead(t, db, key, lorem)
	content := ""
	testWrite(t, db, key, content)
	testRead(t, db, key, content)

	testRead(t, gzipDb, key, content)
	testDelete(t, gzipDb, key)
	testReadEmpty(t, gzipDb, key)
}

func TestScan(t *testing.T) {
	root, err := ioutil.TempDir("", "fsdb_")
	if err != nil {
		t.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)
	opts := local.NewDefaultOptions(root)
	db := local.Open(opts)

	keys := make(map[string]bool)
	keyFunc := func(ret bool) func(key fsdb.Key) bool {
		return func(key fsdb.Key) bool {
			keys[string(key)] = true
			return ret
		}
	}
	err = db.ScanKeys(keyFunc(true), fsdb.IgnoreAllErrFunc)
	if err != nil {
		t.Fatalf("ScanKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Scan empty db got keys: %+v", keys)
	}

	expectKeys := map[string]bool{
		"foo":    true,
		"bar":    true,
		"foobar": true,
	}
	for key := range expectKeys {
		if err := db.Write(fsdb.Key(key), strings.NewReader("")); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	if err := db.ScanKeys(keyFunc(true), fsdb.StopAllErrFunc); err != nil {
		t.Fatalf("ScanKeys failed: %v", err)
	}
	if !reflect.DeepEqual(keys, expectKeys) {
		t.Errorf("ScanKeys expected %+v, got %+v", expectKeys, keys)
	}

	keys = make(map[string]bool)
	if err := db.ScanKeys(keyFunc(false), fsdb.StopAllErrFunc); err != nil {
		t.Fatalf("ScanKeys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("Scan should stop after the first key, got: %+v", keys)
	}
}

func BenchmarkReadWrite(b *testing.B) {
	root, err := ioutil.TempDir(".", "_fsdb_bench_test_")
	if err != nil {
		b.Fatalf("failed to get tmp dir: %v", err)
	}
	defer os.RemoveAll(root)

	keySize := 12
	r := rand.New(rand.NewSource(time.Now().Unix()))

	var benchmarkSizes = map[string]int{
		"1K":   1024,
		"10K":  10 * 1024,
		"1M":   1024 * 1024,
		"10M":  10 * 1024 * 1024,
		"256M": 256 * 1024 * 1024,
	}

	var options = map[string]local.Options{
		"nocompression": local.NewDefaultOptions(root).SetUseGzip(false),
		"gzip-min":      local.NewDefaultOptions(root).SetUseGzip(false).SetGzipLevel(gzip.BestSpeed),
		"gzip-default":  local.NewDefaultOptions(root).SetUseGzip(false).SetGzipLevel(gzip.DefaultCompression),
		"gzip-max":      local.NewDefaultOptions(root).SetUseGzip(false).SetGzipLevel(gzip.BestCompression),
	}

	for label, size := range benchmarkSizes {
		b.Run(
			label,
			func(b *testing.B) {
				content := randomBytes(b, r, size)

				for label, opts := range options {
					b.Run(
						label,
						func(b *testing.B) {
							os.RemoveAll(root)
							db := local.Open(opts)
							keys := make([]fsdb.Key, 0)
							b.Run(
								"write",
								func(b *testing.B) {
									for i := 0; i < b.N; i++ {
										key := fsdb.Key(randomBytes(b, r, keySize))
										keys = append(keys, key)

										err := db.Write(key, bytes.NewReader(content))
										if err != nil {
											b.Fatalf("Write failed: %v", err)
										}
									}
								},
							)
							b.Run(
								"read",
								func(b *testing.B) {
									for i := 0; i < b.N; i++ {
										key := keys[r.Int31n(int32(len(keys)))]
										reader, err := db.Read(key)
										if err != nil {
											b.Fatalf("Read failed: %v", err)
										}
										reader.Close()
									}
								},
							)
						},
					)
				}
			},
		)
	}
}

func randomBytes(b *testing.B, r *rand.Rand, size int) []byte {
	b.Helper()

	reader := io.LimitReader(r, int64(size))
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		b.Fatalf("Generate content failed: %v", err)
	}
	if len(content) != size {
		b.Fatalf(
			"Generate content failed, expected %d bytes, got %d",
			size,
			len(content),
		)
	}
	return content
}

func testDeleteEmpty(t *testing.T, db fsdb.FSDB, key fsdb.Key) {
	t.Helper()
	if err := db.Delete(key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf("Expected NoSuchKeyError, got: %v", err)
	}
}

func testDelete(t *testing.T, db fsdb.FSDB, key fsdb.Key) {
	t.Helper()
	if err := db.Delete(key); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

func testReadEmpty(t *testing.T, db fsdb.FSDB, key fsdb.Key) {
	t.Helper()
	if _, err := db.Read(key); !fsdb.IsNoSuchKeyError(err) {
		t.Errorf("Expected NoSuchKeyError, got: %v", err)
	}
}

func testRead(t *testing.T, db fsdb.FSDB, key fsdb.Key, expect string) {
	t.Helper()
	reader, err := db.Read(key)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	defer reader.Close()
	actual, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("Read content failed: %v", err)
	}
	if string(actual) != expect {
		t.Errorf("Read content expected %q, got %q", expect, actual)
	}
}

func testWrite(t *testing.T, db fsdb.FSDB, key fsdb.Key, data string) {
	t.Helper()
	if err := db.Write(key, strings.NewReader(data)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}
