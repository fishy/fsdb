package local

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/snappy"

	"github.com/fishy/fsdb/interface"
	"github.com/fishy/fsdb/wrapreader"
)

// Make sure the classes satisify interfaces.
var _ fsdb.Local = new(impl)
var _ error = new(KeyCollisionError)

const tempDirPrefix = "fsdb_"
const tempDirMode os.FileMode = 0700

// Filenames used under the entry directory.
const (
	KeyFilename = "key"

	DataFilename       = "data"
	GzipDataFilename   = "data.gz"
	SnappyDataFilename = "data.snappy"
)

// Permissions for files and directories.
var (
	FileModeForFiles os.FileMode = 0600
	FileModeForDirs  os.FileMode = 0700
)

// KeyCollisionError is an error returned when two keys have the same hash.
type KeyCollisionError struct {
	NewKey fsdb.Key
	OldKey fsdb.Key
}

func (err *KeyCollisionError) Error() string {
	return fmt.Sprintf(
		"key collision detected: new key is %q, old key was %q",
		err.NewKey,
		err.OldKey,
	)
}

type impl struct {
	opts Options
}

// Open opens an fsdb with the given options.
//
// There's no need to close it.
func Open(opts Options) fsdb.Local {
	return &impl{
		opts: opts,
	}
}

func (db *impl) Read(key fsdb.Key) (io.ReadCloser, error) {
	dir := db.opts.GetDirForKey(key)
	keyFile := dir + KeyFilename
	if _, err := os.Lstat(keyFile); os.IsNotExist(err) {
		return nil, &fsdb.NoSuchKeyError{Key: key}
	}
	if err := checkKeyCollision(key, keyFile); err != nil {
		return nil, err
	}

	dataFile := dir + DataFilename
	if _, err := os.Lstat(dataFile); err == nil {
		return os.Open(dataFile)
	}

	dataFile = dir + GzipDataFilename
	if _, err := os.Lstat(dataFile); err == nil {
		file, err := os.Open(dataFile)
		if err != nil {
			return nil, err
		}
		reader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		return wrapreader.Wrap(reader, file), nil
	}

	dataFile = dir + SnappyDataFilename
	if _, err := os.Lstat(dataFile); err == nil {
		file, err := os.Open(dataFile)
		if err != nil {
			return nil, err
		}
		return wrapreader.Wrap(snappy.NewReader(file), file), nil
	}

	// Key file exists but there's no data file,
	return nil, &fsdb.NoSuchKeyError{Key: key}
}

func (db *impl) Write(key fsdb.Key, data io.Reader) (err error) {
	dir := db.opts.GetDirForKey(key)
	keyFile := dir + KeyFilename
	if _, err = os.Lstat(keyFile); err == nil {
		if err = checkKeyCollision(key, keyFile); err != nil {
			return err
		}
	}
	tmpdir, err := db.GetTempDir(tempDirPrefix)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	// Write temp key file
	tmpKeyFile := tmpdir + KeyFilename
	if err = func() error {
		f, err := createFile(tmpKeyFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = io.Copy(f, bytes.NewReader(key)); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	// Write temp data file
	var tmpDataFile string
	var dataFile string
	if db.opts.GetUseSnappy() {
		tmpDataFile = tmpdir + SnappyDataFilename
		dataFile = dir + SnappyDataFilename
		if err = func() error {
			f, err := createFile(tmpDataFile)
			if err != nil {
				return err
			}
			defer f.Close()
			writer := snappy.NewBufferedWriter(f)
			defer writer.Close()
			if _, err = io.Copy(writer, data); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	} else if db.opts.GetUseGzip() {
		tmpDataFile = tmpdir + GzipDataFilename
		dataFile = dir + GzipDataFilename
		if err = func() error {
			f, err := createFile(tmpDataFile)
			if err != nil {
				return err
			}
			defer f.Close()
			writer, err := gzip.NewWriterLevel(f, db.opts.GetGzipLevel())
			if err != nil {
				return err
			}
			defer writer.Close()
			if _, err = io.Copy(writer, data); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	} else {
		tmpDataFile = tmpdir + DataFilename
		dataFile = dir + DataFilename
		if err = func() error {
			f, err := createFile(tmpDataFile)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err = io.Copy(f, data); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	// Move data file
	if err = os.MkdirAll(dir, FileModeForDirs); err != nil && !os.IsExist(err) {
		return err
	}
	for _, file := range []string{
		DataFilename,
		SnappyDataFilename,
		GzipDataFilename,
	} {
		if err = os.Remove(dir + file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err = os.Rename(tmpDataFile, dataFile); err != nil {
		return err
	}

	// Move key file
	if err = os.Rename(tmpKeyFile, keyFile); err != nil {
		return err
	}
	return nil
}

func (db *impl) Delete(key fsdb.Key) error {
	dir := db.opts.GetDirForKey(key)
	keyFile := dir + KeyFilename
	if _, err := os.Lstat(keyFile); os.IsNotExist(err) {
		return &fsdb.NoSuchKeyError{Key: key}
	}
	if err := checkKeyCollision(key, keyFile); err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (db *impl) GetRootDataDir() string {
	return db.opts.GetDataDir()
}

func (db *impl) GetTempDir(prefix string) (dir string, err error) {
	root := db.opts.GetTempDir()
	if err = os.MkdirAll(root, tempDirMode); err != nil && !os.IsExist(err) {
		return
	}
	dir, err = ioutil.TempDir(db.opts.GetTempDir(), prefix)
	if !strings.HasSuffix(dir, PathSeparator) {
		dir += PathSeparator
	}
	return
}

func (db *impl) ScanKeys(
	keyFunc func(key fsdb.Key) bool,
	errFunc func(err error) bool,
) error {
	_, err := scanKeys(db.GetRootDataDir(), keyFunc, errFunc)
	return err
}

func scanKeys(
	root string,
	keyFunc func(key fsdb.Key) bool,
	errFunc func(err error) bool,
) (bool, error) {
	dir, err := os.Open(root)
	if err != nil {
		if errFunc == nil || !errFunc(err) {
			return false, err
		}
	}
	infos, err := dir.Readdir(-1)
	dir.Close()
	if err != nil {
		if errFunc == nil || !errFunc(err) {
			return false, err
		}
	}
	if len(infos) == 0 && err == nil {
		// Empty direcoty, do some cleanup here.
		os.Remove(root)
		return true, nil
	}
	for _, info := range infos {
		if info.IsDir() {
			ret, err := scanKeys(root+info.Name()+PathSeparator, keyFunc, errFunc)
			if err != nil {
				return false, err
			}
			if !ret {
				return ret, nil
			}
			continue
		}
		if info.Name() == KeyFilename {
			path := root + info.Name()
			key, err := readKey(path)
			if err != nil {
				if errFunc == nil || !errFunc(err) {
					return false, err
				}
			}
			ret := keyFunc(key)
			if !ret {
				return ret, nil
			}
		}
	}
	return true, nil
}

func checkKeyCollision(key fsdb.Key, path string) error {
	old, err := readKey(path)
	if err != nil {
		return err
	}
	if key.Equals(old) {
		return nil
	}
	return &KeyCollisionError{
		NewKey: key,
		OldKey: old,
	}
}

func readKey(path string) (fsdb.Key, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	key, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return fsdb.Key(key), nil
}

func createFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, FileModeForFiles)
}
