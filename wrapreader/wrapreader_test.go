package wrapreader_test

import (
	"testing"

	"github.com/fishy/fsdb/wrapreader"
)

type DummyReader struct {
	ReadCalled bool
}

func (reader *DummyReader) Read(p []byte) (n int, err error) {
	reader.ReadCalled = true
	return
}

type DummyReadCloser struct {
	ReadCalled  bool
	CloseCalled bool
}

func (closer *DummyReadCloser) Read(p []byte) (n int, err error) {
	closer.ReadCalled = true
	return
}

func (closer *DummyReadCloser) Close() (err error) {
	closer.CloseCalled = true
	return
}

func TestRead(t *testing.T) {
	reader := new(DummyReader)
	closer := new(DummyReadCloser)
	wrap := wrapreader.Wrap(reader, closer)
	wrap.Read(nil)
	if !reader.ReadCalled {
		t.Error("WrapReader.Read should call the underlying reader's Read function")
	}
	if closer.ReadCalled {
		t.Error(
			"WrapReader.Read should not call the underlying closer's Read function",
		)
	}
}

func TestClose(t *testing.T) {
	reader := new(DummyReader)
	closer := new(DummyReadCloser)
	wrap := wrapreader.Wrap(reader, closer)
	wrap.Close()
	if !closer.CloseCalled {
		t.Error(
			"WrapReader.Close should call the underlying closer's Close function",
		)
	}
}

func TestCloseBoth(t *testing.T) {
	reader := new(DummyReadCloser)
	closer := new(DummyReadCloser)
	wrap := wrapreader.Wrap(reader, closer)
	wrap.Close()
	if !closer.CloseCalled {
		t.Error(
			"WrapReader.Close should call the underlying closer's Close function",
		)
	}
	if !reader.CloseCalled {
		t.Error(
			"WrapReader.Close should call the underlying reader's Close function",
		)
	}
}
