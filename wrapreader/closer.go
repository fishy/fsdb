package wrapreader

import (
	"io"
)

type dummyCloser []byte

func (c dummyCloser) Close() error {
	return nil
}

// ReaderToReadCloser converts an io.Reader into io.ReadCloser.
func ReaderToReadCloser(reader io.Reader) io.ReadCloser {
	var c dummyCloser
	return Wrap(reader, c)
}
