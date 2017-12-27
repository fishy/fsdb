package wrapreader

import (
	"io"

	"github.com/fishy/fsdb/errbatch"
)

type wrapReader struct {
	reader io.Reader
	closer io.Closer
}

// Wrap wraps reader and closer together to create an new io.ReadCloser.
//
// The Read function will simply call the wrapped reader's Read function,
// while the Close function will call the wrapped closer's Close function.
//
// If the wrapped reader is also an io.Closer,
// its Close function will be called in Close as well.
func Wrap(reader io.Reader, closer io.Closer) io.ReadCloser {
	return &wrapReader{
		reader: reader,
		closer: closer,
	}
}

func (r *wrapReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *wrapReader) Close() error {
	err := errbatch.NewErrBatch()
	if closer, ok := r.reader.(io.Closer); ok {
		err.Add(closer.Close())
	}
	err.Add(r.closer.Close())
	return err.Compile()
}
