package bucket

import (
	"io"
)

// Bucket defines the interface for a remote storage bucket (e.g. s3 or gcs).
type Bucket interface {
	// Read downloads an entry from the bucket.
	//
	// It's the caller's responsibility to close the ReadCloser returned.
	Read(name string) (io.ReadCloser, error)

	// Write uploads an entry to the bucket.
	Write(name string, data io.Reader) error

	// Delete deletes an entry from the bucket.
	Delete(name string) error

	// IsNotExist checks wether an error returned by Read or Delete means the
	// entry does not exist on the bucket.
	IsNotExist(err error) bool
}
