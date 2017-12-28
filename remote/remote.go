package remote

import (
	"bytes"
	"compress/gzip"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fishy/fsdb/bucket"
	"github.com/fishy/fsdb/errbatch"
	"github.com/fishy/fsdb/interface"
)

const tempDirPrefix = "remote_"
const tempFilename = "data"

var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

type remoteDB struct {
	local  fsdb.Local
	bucket bucket.Bucket
	opts   Options
}

// Open creates a remote FSDB,
// which is backed by a local FSDB and a remote bucket.
//
// There's no need to close.
//
// Read reads from local first,
// then read from remote bucket if it does not exist locally.
// In that case,
// the data will be saved locally for cache until the next upload loop.
//
// Write writes locally.
// There is a background scan loop to upload everything from local to remote,
// then deletes the local copy after the upload succeed.
//
// Delete deletes from both local and remote,
// and returns combined errors, if any.
func Open(local fsdb.Local, bucket bucket.Bucket, opts Options) fsdb.FSDB {
	db := &remoteDB{
		local:  local,
		bucket: bucket,
		opts:   opts,
	}
	go db.startScanLoop()
	return db
}

func (db *remoteDB) Read(key fsdb.Key) (data io.ReadCloser, err error) {
	data, err = db.local.Read(key)
	if err == nil {
		return data, nil
	}
	if !fsdb.IsNoSuchKeyError(err) {
		return nil, err
	}
	if logger := db.opts.GetLogger(); logger != nil {
		started := time.Now()
		defer logger.Printf(
			"download %v from bucket took %v, error: %v",
			key,
			time.Now().Sub(started),
			err,
		)
	}
	remoteData, err := db.bucket.Read(db.opts.GetRemoteName(key))
	if remoteData != nil {
		defer remoteData.Close()
	}
	if !db.bucket.IsNotExist(err) {
		if err != nil {
			return nil, err
		}
		buf, _ := ioutil.ReadAll(remoteData)
		gzipReader, err := gzip.NewReader(bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		tmpdir, tmpfile, err := db.saveTempFile(gzipReader)
		if tmpdir != "" {
			defer os.Remove(tmpdir)
		}
		if tmpfile != "" {
			defer os.Remove(tmpfile)
		}
		if err != nil {
			return nil, err
		}
		// Read from local again, so that in case a new write happened during
		// downloading, we don't overwrite it with stale remote data.
		data, err = db.local.Read(key)
		if err == nil {
			return data, nil
		}
		f, err := os.Open(tmpfile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err = db.local.Write(key, f); err != nil {
			return nil, err
		}
	}
	return db.local.Read(key)
}

func (db *remoteDB) Delete(key fsdb.Key) error {
	existNeither := true

	ret := errbatch.NewErrBatch()
	err := db.local.Delete(key)
	if !fsdb.IsNoSuchKeyError(err) {
		existNeither = false
		ret.Add(err)
	}
	err = db.bucket.Delete(db.opts.GetRemoteName(key))
	if !db.bucket.IsNotExist(err) {
		existNeither = false
		ret.Add(err)
	}

	if existNeither {
		return &fsdb.NoSuchKeyError{Key: key}
	}
	return ret.Compile()
}

func (db *remoteDB) Write(key fsdb.Key, data io.Reader) error {
	return db.local.Write(key, data)
}

func (db *remoteDB) readAndCRC(key fsdb.Key) (uint32, []byte, error) {
	reader, err := db.local.Read(key)
	if err != nil {
		return 0, nil, err
	}
	defer reader.Close()
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, nil, err
	}
	return crc32.Checksum(buf, crc32cTable), buf, nil
}

func (db *remoteDB) saveTempFile(data io.Reader) (
	dir string,
	file string,
	err error,
) {
	dir, err = db.local.GetTempDir(tempDirPrefix)
	if err != nil {
		return
	}
	file = dir + tempFilename
	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, data)
	return
}

func (db *remoteDB) uploadKey(key fsdb.Key) error {
	oldCrc, content, err := db.readAndCRC(key)
	if err != nil {
		return err
	}
	reader, err := gzipData(bytes.NewReader(content))
	if err != nil {
		return err
	}
	if err = db.bucket.Write(db.opts.GetRemoteName(key), reader); err != nil {
		return err
	}
	// check crc again before deleting
	newCrc, _, err := db.readAndCRC(key)
	if err != nil {
		return err
	}
	if newCrc == oldCrc {
		return db.local.Delete(key)
	}
	return nil
}

func (db *remoteDB) startScanLoop() {
	for range time.Tick(db.opts.GetUploadDelay()) {
		db.scanLoop()
	}
}

func (db *remoteDB) scanLoop() {
	n := db.opts.GetUploadThreadNum()
	logger := db.opts.GetLogger()
	keyChan := make(chan fsdb.Key, 0)

	scanned := initAtomicInt64()
	skipped := initAtomicInt64()
	uploaded := initAtomicInt64()
	failed := initAtomicInt64()

	var wg sync.WaitGroup
	wg.Add(n)

	// Workers
	for i := 0; i < n; i++ {
		go func() {
			for key := range keyChan {
				atomic.AddInt64(scanned, 1)
				if db.opts.SkipKey(key) {
					atomic.AddInt64(skipped, 1)
					continue
				}
				if err := db.uploadKey(key); err != nil {
					// All errors will be retried on next scan loop, just log and ignore.
					if logger != nil {
						logger.Printf("failed to upload %v to bucket: %v", key, err)
					}
					atomic.AddInt64(failed, 1)
				} else {
					atomic.AddInt64(uploaded, 1)
				}
			}
			wg.Done()
		}()
	}

	started := time.Now()
	if err := db.local.ScanKeys(
		func(key fsdb.Key) bool {
			keyChan <- key
			return true
		},
		func(err error) bool {
			// Most I/O errors here are just caused by race conditions,
			// safe to log and ignore.
			if logger != nil {
				logger.Printf("ScanKeys reported error: %v", err)
			}
			return true
		},
	); err != nil {
		if logger != nil {
			logger.Printf("ScanKeys returned error: %v", err)
		}
	} else {
		close(keyChan)
		wg.Wait()
		if logger != nil {
			logger.Printf(
				"took %v, scanned %d, skipped %d, uploaded %d, failed %d",
				time.Now().Sub(started),
				atomic.LoadInt64(scanned),
				atomic.LoadInt64(skipped),
				atomic.LoadInt64(uploaded),
				atomic.LoadInt64(failed),
			)
		}
	}
}

func initAtomicInt64() *int64 {
	ret := new(int64)
	atomic.StoreInt64(ret, 0)
	return ret
}

func gzipData(data io.Reader) (io.Reader, error) {
	buf := new(bytes.Buffer)
	writer, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	if _, err = io.Copy(writer, data); err != nil {
		return nil, err
	}
	return buf, nil
}
