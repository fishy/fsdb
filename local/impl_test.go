package local_test

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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
	root := os.TempDir() + local.PathSeparator + "fsdb"
	defer os.RemoveAll(root)
	opts := local.NewDefaultOptions(root).SetUseGzip(false).SetUseSnappy(false)
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

func TestSnappy(t *testing.T) {
	root := os.TempDir() + local.PathSeparator + "fsdb"
	defer os.RemoveAll(root)
	opts := local.NewDefaultOptions(root).SetUseSnappy(true)
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
	root := os.TempDir() + local.PathSeparator + "fsdb"
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
	root := os.TempDir() + local.PathSeparator + "fsdb"
	defer os.RemoveAll(root)
	snappyOpts := local.NewDefaultOptions(root).SetUseSnappy(true)
	snappyDb := local.Open(snappyOpts)

	key := fsdb.Key("foo")
	testWrite(t, snappyDb, key, lorem)
	testRead(t, snappyDb, key, lorem)

	gzipOpts := local.NewDefaultOptions(root).SetUseGzip(true)
	gzipDb := local.Open(gzipOpts)
	testRead(t, gzipDb, key, lorem)
	content := ""
	testWrite(t, gzipDb, key, content)
	testRead(t, gzipDb, key, content)

	opts := local.NewDefaultOptions(root).SetUseGzip(false).SetUseSnappy(false)
	db := local.Open(opts)
	testRead(t, db, key, content)
	testDelete(t, db, key)
	testReadEmpty(t, db, key)
}

func TestDirs(t *testing.T) {
	root := os.TempDir() + local.PathSeparator + "fsdb"
	defer os.RemoveAll(root)
	opts := local.NewDefaultOptions(root)
	db := local.Open(opts)

	expect := opts.GetDataDir()
	actual := db.GetRootDataDir()
	if expect != actual {
		t.Errorf("GetRootDataDir() expected %q, got %q", expect, actual)
	}

	expect = opts.GetTempDir()
	actual, err := db.GetTempDir()
	if err != nil {
		t.Fatalf("GetTempDir() failed: %v", err)
	}
	if !strings.HasPrefix(actual, expect) {
		t.Errorf("GetTempDir() should be under %q, got %q", expect, actual)
	}
	if _, err := os.Lstat(actual); err != nil {
		t.Errorf("%q should exists, got: %v", actual, err)
	}
}

func BenchmarkReadWrite(b *testing.B) {
	root := "_fsdb_tmp"
	key := fsdb.Key("foo")

	var benchmarkSizes = map[string]int{
		"1K":   1024,
		"10K":  10 * 1024,
		"1M":   1024 * 1024,
		"10M":  10 * 1024 * 1024,
		"256M": 256 * 1024 * 1024,
	}

	var options = map[string]local.Options{
		"nocompression": local.NewDefaultOptions(root).SetUseGzip(false).SetUseSnappy(false),
		"snappy":        local.NewDefaultOptions(root).SetUseSnappy(true),
		"gzip-min":      local.NewDefaultOptions(root).SetUseGzip(false).SetGzipLevel(gzip.BestSpeed),
		"gzip-default":  local.NewDefaultOptions(root).SetUseGzip(false).SetGzipLevel(gzip.DefaultCompression),
		"gzip-max":      local.NewDefaultOptions(root).SetUseGzip(false).SetGzipLevel(gzip.BestCompression),
	}

	for label, size := range benchmarkSizes {
		b.Run(
			label,
			func(b *testing.B) {
				randReader := io.LimitReader(rand.Reader, int64(size))
				content, err := ioutil.ReadAll(randReader)
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

				for label, opts := range options {
					b.Run(
						label,
						func(b *testing.B) {
							defer os.RemoveAll(root)
							db := local.Open(opts)
							b.Run(
								"write",
								func(b *testing.B) {
									for i := 0; i < b.N; i++ {
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
