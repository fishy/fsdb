package fsdb

import (
	"io"
)

// FSDB defines the interface for an FSDB implementation.
type FSDB interface {
	// Read opens an entry and returns a ReadCloser.
	//
	// If the key does not exist, it should return a NoSuchKeyError.
	//
	// It should never return both nil reader and nil err.
	//
	// It's the caller's responsibility to close the ReadCloser returned.
	Read(key Key) (reader io.ReadCloser, err error)

	// Write opens an entry and returns a WriteCloser.
	//
	// If the key already exists, it will be overwritten.
	//
	// If data is actually a ReadCloser,
	// it's the caller's responsibility to close it after Write function returns.
	Write(key Key, data io.Reader) error

	// Delete deletes an entry.
	//
	// If the key does not exist, it should return a NoSuchKeyError.
	Delete(key Key) error
}

// Local defines extra interface for a local FSDB implementation.
type Local interface {
	FSDB

	// ScanKeys scans all the keys locally.
	//
	// This function would be heavy on IO and takes a long time. Use with caution.
	//
	// The behavior is undefined for keys changed after the scan started.
	ScanKeys(keyFunc KeyFunc, errFunc ErrFunc) error
}

// KeyFunc is used in ScanKeys function in Local interface.
//
// It's the callback function called for every key scanned.
//
// It should return true to continue the scan and false to abort the scan.
//
// It's OK for KeyFunc to block.
type KeyFunc func(key Key) bool

// ErrFunc is used in ScanKeys function in Local interface.
//
// It's the callback function called when the scan encounters an I/O error that
// is possible to be ignored.
type ErrFunc func(err error) bool

// StopAllErrFunc is an ErrFunc that can be used in Local.ScanKeys().
//
// It always returns false,
// means the scan stops at the first I/O errors it encounters.
func StopAllErrFunc(err error) bool {
	return false
}

// IgnoreAllErrFunc is an ErrFunc that can be used in Local.ScanKeys().
//
// It always returns true, means the scan ignores all I/O errors if possible.
func IgnoreAllErrFunc(err error) bool {
	return true
}
