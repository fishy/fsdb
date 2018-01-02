package hybrid

import (
	"bytes"
	"compress/gzip"
	"context"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"sync/atomic"
	"time"

	"github.com/fishy/fsdb/bucket"
	"github.com/fishy/fsdb/fsdb"
	"github.com/fishy/fsdb/libs/errbatch"
	"github.com/fishy/fsdb/libs/rowlock"
)

const tempFilename = "data"

var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

type impl struct {
	local  fsdb.Local
	bucket bucket.Bucket
	opts   Options
	locks  *rowlock.RowLock
}

// Open creates a hybrid FSDB,
// which is backed by a local FSDB and a remote bucket.
//
// There's no need to close,
// but you could cancel the context to stop the upload loop.
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
//
// github.com/fishy/gcsbucket and github.com/fishy/s3bucket provide
// bucket.Bucket implementations for Google Cloud Storage and AWS S3,
// respectively.
func Open(
	ctx context.Context,
	local fsdb.Local,
	bucket bucket.Bucket,
	opts Options,
) fsdb.FSDB {
	db := &impl{
		local:  local,
		bucket: bucket,
		opts:   opts,
		locks:  rowlock.NewRowLock(rowlock.MutexNewLocker),
	}
	go db.startScanLoop(ctx)
	return db
}

func (db *impl) Read(ctx context.Context, key fsdb.Key) (io.ReadCloser, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := db.local.Read(ctx, key)
	if err == nil {
		return data, nil
	}
	if !fsdb.IsNoSuchKeyError(err) {
		return nil, err
	}
	remoteData, err := db.readBucket(ctx, key)
	if !db.bucket.IsNotExist(err) {
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		if db.opts.GetUseLock() {
			db.locks.Lock(string(key))
			defer db.locks.Unlock(string(key))
		}
		// Read from local again, so that in case a new write happened during
		// downloading, we don't overwrite it with stale remote data.
		data, err = db.local.Read(ctx, key)
		if err == nil {
			return data, nil
		}
		if err := db.local.Write(ctx, key, remoteData); err != nil {
			return nil, err
		}
	}
	return db.local.Read(ctx, key)
}

func (db *impl) Write(ctx context.Context, key fsdb.Key, data io.Reader) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	if db.opts.GetUseLock() {
		db.locks.Lock(string(key))
		defer db.locks.Unlock(string(key))
	}
	return db.local.Write(ctx, key, data)
}

func (db *impl) Delete(ctx context.Context, key fsdb.Key) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	existNeither := true

	ret := errbatch.NewErrBatch()
	err := db.local.Delete(ctx, key)
	if !fsdb.IsNoSuchKeyError(err) {
		existNeither = false
		ret.Add(err)
	}
	err = db.bucket.Delete(ctx, db.opts.GetRemoteName(key))
	if !db.bucket.IsNotExist(err) {
		existNeither = false
		ret.Add(err)
	}

	if existNeither {
		return &fsdb.NoSuchKeyError{Key: key}
	}
	return ret.Compile()
}

// readBucket reads the key from remote bucket fully.
func (db *impl) readBucket(
	ctx context.Context,
	key fsdb.Key,
) (io.Reader, error) {
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	started := time.Now()
	data, err := db.bucket.Read(ctx, db.opts.GetRemoteName(key))
	if err != nil {
		return nil, err
	}
	defer data.Close()
	if logger := db.opts.GetLogger(); logger != nil {
		defer logger.Printf(
			"download %v from bucket took %v",
			key,
			time.Now().Sub(started),
		)
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	gzipReader, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	buf, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// readAndCRC reads the key from local fully, and calculates crc32c.
func (db *impl) readAndCRC(
	ctx context.Context,
	key fsdb.Key,
) (uint32, []byte, error) {
	select {
	default:
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	}

	reader, err := db.local.Read(ctx, key)
	if err != nil {
		return 0, nil, err
	}
	defer reader.Close()
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	}

	return crc32.Checksum(buf, crc32cTable), buf, nil
}

// uploadKey uploads a key to remote bucket, and deletes the local copy.
func (db *impl) uploadKey(ctx context.Context, key fsdb.Key) error {
	oldCrc, content, err := db.readAndCRC(ctx, key)
	if err != nil {
		return err
	}
	reader, err := gzipData(bytes.NewReader(content))
	if err != nil {
		return err
	}

	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	err = db.bucket.Write(ctx, db.opts.GetRemoteName(key), reader)
	if err != nil {
		return err
	}

	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	if db.opts.GetUseLock() {
		db.locks.Lock(string(key))
		defer db.locks.Unlock(string(key))
	}
	// check crc again before deleting
	newCrc, _, err := db.readAndCRC(ctx, key)
	if err != nil {
		return err
	}

	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	if newCrc == oldCrc {
		return db.local.Delete(ctx, key)
	}
	return nil
}

func (db *impl) startScanLoop(ctx context.Context) {
	select {
	default:
	case <-ctx.Done():
		return
	}

	n := db.opts.GetUploadThreadNum()
	logger := db.opts.GetLogger()
	keys := make(chan fsdb.Key, 0)

	scanned := new(int64)
	skipped := new(int64)
	uploaded := new(int64)
	failed := new(int64)

	// Workers
	for i := 0; i < n; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case key := <-keys:
					atomic.AddInt64(scanned, 1)
					if db.opts.SkipKey(key) {
						atomic.AddInt64(skipped, 1)
						continue
					}
					if err := db.uploadKey(ctx, key); err != nil {
						// All errors will be retried on next scan loop,
						// safe to just log and ignore.
						if logger != nil {
							logger.Printf("failed to upload %v to bucket: %v", key, err)
						}
						atomic.AddInt64(failed, 1)
					} else {
						atomic.AddInt64(uploaded, 1)
					}
				}
			}
		}()
	}
	ticker := time.NewTicker(db.opts.GetUploadDelay())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			atomic.StoreInt64(scanned, 0)
			atomic.StoreInt64(skipped, 0)
			atomic.StoreInt64(uploaded, 0)
			atomic.StoreInt64(failed, 0)

			started := time.Now()

			if err := db.local.ScanKeys(
				ctx,
				func(key fsdb.Key) bool {
					select {
					case <-ctx.Done():
						return false
					default:
						keys <- key
						return true
					}
				},
				func(path string, err error) bool {
					// Most I/O errors here are just not exist errors caused by race
					// conditions, log if it's not not exist error and ignore.
					if logger != nil && !os.IsNotExist(err) {
						logger.Printf("ScanKeys reported error on %s: %v", path, err)
					}
					return true
				},
			); err != nil {
				if logger != nil {
					logger.Printf("ScanKeys returned error: %v", err)
				}
			}

			if logger != nil {
				// The skipped/uploaded/failed value could be off by less than twice the
				// worker number, as when we print this log the workers are likely not
				// finished with the keys yet, and when we start the next loop the
				// workers might be still working on keys from the previous loop.
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
