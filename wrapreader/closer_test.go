package wrapreader_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/fishy/fsdb/wrapreader"
)

func TestReaderToReadCloser(t *testing.T) {
	reader := new(DummyReadCloser)
	closer := wrapreader.ReaderToReadCloser(reader)
	closer.Read(nil)
	if !reader.ReadCalled {
		t.Error("Converted closer should call underlying reader's Read function.")
	}
	closer.Close()
	if !reader.CloseCalled {
		t.Error(
			"Converted closer should call underlying readCloser's Close function.",
		)
	}
}

func ExampleReaderToReadCloser() {
	// reader only satisifies io.Reader, but not io.ReadCloser
	reader := strings.NewReader("Hello, world!")

	// readCloser is an io.ReadCloser
	readCloser := wrapreader.ReaderToReadCloser(reader)
	defer readCloser.Close()
	content, err := ioutil.ReadAll(readCloser)
	if err != nil {
		// TODO: handle error
	}
	fmt.Println(string(content))
	// Output:
	// Hello, world!
}
