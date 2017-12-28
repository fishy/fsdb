package local_test

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fishy/fsdb/interface"
	"github.com/fishy/fsdb/local"
)

func TestOptions(t *testing.T) {
	root := "/foobar"
	opts := local.NewDefaultOptions(root)

	t.Run(
		"compression",
		func(t *testing.T) {
			opts.SetUseGzip(true)
			if !opts.GetUseGzip() {
				t.Error("gzip should be set to true")
			}

			opts.SetUseSnappy(true)
			if !opts.GetUseSnappy() {
				t.Error("snappy should be set to true")
			}
			if opts.GetUseGzip() {
				t.Error("set snappy should disable gzip")
			}

			opts.SetUseGzip(true)
			if !opts.GetUseGzip() {
				t.Error("gzip should be set to true")
			}
			if opts.GetUseSnappy() {
				t.Error("set gzip should disable snappy")
			}
		},
	)

	t.Run(
		"dirs",
		func(t *testing.T) {
			var expect, actual string

			opts.SetDataDir("data2")
			expect = "/foobar" + local.PathSeparator + "data2" + local.PathSeparator
			actual = opts.GetDataDir()
			if expect != actual {
				t.Errorf("data dir expected %q, got %q", expect, actual)
			}

			opts.SetDataDir("data2" + local.PathSeparator)
			expect = "/foobar" + local.PathSeparator + "data2" + local.PathSeparator
			actual = opts.GetDataDir()
			if expect != actual {
				t.Errorf("data dir expected %q, got %q", expect, actual)
			}

			opts.SetTempDir("tmp")
			expect = "/foobar" + local.PathSeparator + "tmp" + local.PathSeparator
			actual = opts.GetTempDir()
			if expect != actual {
				t.Errorf("data dir expected %q, got %q", expect, actual)
			}

			opts.SetTempDir("tmp" + local.PathSeparator)
			expect = "/foobar" + local.PathSeparator + "tmp" + local.PathSeparator
			actual = opts.GetTempDir()
			if expect != actual {
				t.Errorf("data dir expected %q, got %q", expect, actual)
			}
		},
	)

	t.Run(
		"key-hash",
		func(t *testing.T) {
			key := fsdb.Key("key")
			data := "data"
			opts.SetDataDir(data)
			var expect, actual string

			expect = strings.Join(
				[]string{
					root,
					data,
					"6c",
					"b1",
					"b0",
					"e50d74419e2244eaa7328235f71b48c7e1c33b23f6f9517d14",
					"",
				},
				local.PathSeparator,
			)
			actual = opts.GetDirForKey(key)
			if actual != expect {
				t.Errorf("hash dir for key %q expected %q, got %q", key, expect, actual)
			}

			opts.SetDirLevel(sha512.Size224 + 10)
			expect = strings.Join(
				[]string{
					root,
					data,
					"6c",
					"b1",
					"b0",
					"e5",
					"0d",
					"74",
					"41",
					"9e",
					"22",
					"44",
					"ea",
					"a7",
					"32",
					"82",
					"35",
					"f7",
					"1b",
					"48",
					"c7",
					"e1",
					"c3",
					"3b",
					"23",
					"f6",
					"f9",
					"51",
					"7d",
					"14",
					"",
				},
				local.PathSeparator,
			)
			actual = opts.GetDirForKey(key)
			if actual != expect {
				t.Errorf("hash dir for key %v expected %q, got %q", key, expect, actual)
			}

			opts.SetDirLevel(sha512.Size224)
			actual = opts.GetDirForKey(key)
			if actual != expect {
				t.Errorf("hash dir for key %v expected %q, got %q", key, expect, actual)
			}
		},
	)

	t.Run(
		"hash-reentrant",
		func(t *testing.T) {
			if testing.Short() {
				t.Skip("skipping test in short mode")
			}

			calcHash := func(h hash.Hash, key fsdb.Key, sleep time.Duration) string {
				h.Write(key)
				time.Sleep(sleep)
				return hex.EncodeToString(h.Sum([]byte{}))
			}

			keys := []fsdb.Key{
				fsdb.Key("foo"),
				fsdb.Key("bar"),
				fsdb.Key("key"),
			}
			expect := make([]string, len(keys))
			for i, key := range keys {
				expect[i] = calcHash(sha512.New512_224(), key, 0)
			}

			opts.SetHashFunc(sha512.New512_224)
			var wg sync.WaitGroup
			wg.Add(len(keys))
			sleep := time.Millisecond * 100
			for i, key := range keys {
				go func(key fsdb.Key, expect string) {
					defer wg.Done()

					actual := calcHash(opts.GetHashFunc()(), key, sleep)
					if actual != expect {
						t.Errorf("hash for %v expected %q, got %q", key, expect, actual)
					}
				}(key, expect[i])
			}
			wg.Wait()
		},
	)
}
