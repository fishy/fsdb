// Package wrapreader provides an io.ReadCloser that wraps an io.Reader and
// an io.Closer together.
//
// It's useful when dealing with io.Reader implementations that wraps another
// io.ReadCloser, but will not close the underlying reader, such as GzipReader.
//
// It also provides a ReaderToReadCloser function to convert an io.Reader into
// io.ReadCloser.
package wrapreader
