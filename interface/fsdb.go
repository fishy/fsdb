package fsdb

import (
	"io"
)

// FSDB defines the interface for an FSDB implementation.
type FSDB interface {
	// Read opens an entry and returns a ReadCloser.
	//
	// If the key does not exist, it should return an error of ErrNoSuchKey.
	//
	// It should never return both nil reader and nil err.
	//
	// It's the caller's responsibility to close the ReadCloser returned.
	Read(key Key) (reader io.ReadCloser, err error)

	// Write opens an entry and returns a WriteCloser.
	//
	// If the key already exists, it will be overwritten.
	//
	// If data is actually a ReadCloser, it's the caller's responsibility to close
	// it after Write function returns.
	Write(key Key, data io.Reader) error

	// Delete deletes an entry.
	//
	// If the key does not exist, it should return an error of ErrNoSuchKey.
	Delete(key Key) error
}

// Local defines extra interface for a local FSDB implementation.
type Local interface {
	FSDB

	// GetRootDataDir returns the root data directory of the implementation.
	//
	// It's guaranteed to end with os.PathSeparator.
	GetRootDataDir() string

	// GetTempDir returns a temporary directory that's guaranteed to be on the
	// same file system of the data directory.
	//
	// It's guaranteed to end with os.PathSeparator.
	//
	// It's the caller's responsibility to delete the directory after use.
	GetTempDir() string

	// ScanKeys scans all the keys locally.
	//
	// The keyFunc parameter is the callback function called for every key it
	// scanned. It should return true to continue to scan and false to abort the
	// scan. The keyFunc must not be nil.
	//
	// The errFunc parameter is the callback function called when the scan
	// encounters an error. It should return true if the error is safe to ignore
	// and continue the scan. The errFunc could be nil, which means the scan stops
	// at the first error it encounters.
	//
	// This function would be heavy on IO and takes a long time. Use with caution.
	//
	// The behavior is undefined for keys changed after the scan started.
	ScanKeys(keyFunc func(key Key) bool, errFunc func(err error) bool) error
}
