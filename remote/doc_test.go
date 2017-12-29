package remote_test

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fishy/fsdb/bucket"
	"github.com/fishy/fsdb/interface"
	"github.com/fishy/fsdb/local"
	"github.com/fishy/fsdb/remote"
)

func Example() {
	localRoot, _ := ioutil.TempDir("", "fsdb_")
	defer os.RemoveAll(localRoot)

	var bucket bucket.Bucket
	// TODO: open bucket from an implementation

	ctx, cancel := context.WithCancel(context.Background())
	db := remote.Open(
		ctx,
		local.Open(local.NewDefaultOptions(localRoot)),
		bucket,
		remote.NewDefaultOptions(),
	)
	defer cancel() // Stop the upload loop, not really necessary

	key := fsdb.Key("key")

	if err := db.Write(ctx, key, strings.NewReader("Hello, world!")); err != nil {
		// TODO: handle error
	}

	reader, err := db.Read(ctx, key)
	if err != nil {
		// TODO: handle error
	}
	defer reader.Close()
	// TODO: read from reader

	if err := db.Delete(ctx, key); err != nil {
		// TODO: handle error
	}
}
