package fsdb_test

import (
	"context"
	"strings"

	"github.com/fishy/fsdb"
)

func Example() {
	key := fsdb.Key("key")
	ctx := context.Background()
	var db fsdb.Local
	// TODO: open from an implementation

	if err := db.Write(ctx, key, strings.NewReader("content")); err != nil {
		// TODO: handle error
	}

	reader, err := db.Read(ctx, key)
	if err != nil {
		// TODO: handle error
	}
	defer reader.Close()
	// TODO: read from reader

	if err := db.ScanKeys(
		ctx,
		func(key fsdb.Key) bool {
			// TODO: emit the key
			return true // return true to continue the scan
		},
		fsdb.StopAll,
	); err != nil {
		// TODO: handle error
	}

	if err := db.Delete(ctx, key); err != nil {
		// TODO: handle error
	}
}
