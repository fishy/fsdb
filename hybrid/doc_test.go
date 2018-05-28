package hybrid_test

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fishy/fsdb"
	"github.com/fishy/fsdb/bucket"
	"github.com/fishy/fsdb/hybrid"
	"github.com/fishy/fsdb/local"
)

func Example() {
	root, _ := ioutil.TempDir("", "fsdb_")
	defer os.RemoveAll(root)

	var bucket bucket.Bucket
	// TODO: open bucket from an implementation

	ctx, cancel := context.WithCancel(context.Background())
	db := hybrid.Open(
		ctx,
		local.Open(local.NewDefaultOptions(root)),
		bucket,
		hybrid.NewDefaultOptions(),
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
