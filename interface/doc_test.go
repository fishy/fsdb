package fsdb_test

import (
	"strings"

	"github.com/fishy/fsdb/interface"
)

func Example() {
	key := fsdb.Key("key")
	var db fsdb.Local
	// TODO: open from an implementation

	if err := db.Write(key, strings.NewReader("content")); err != nil {
		// TODO: handle error
	}

	reader, err := db.Read(key)
	if err != nil {
		// TODO: handle error
	}
	defer reader.Close()
	// TODO: read from reader

	if err := db.ScanKeys(
		func(key fsdb.Key) bool {
			// TODO: emit the key
			return true // return true to continue the scan
		},
		fsdb.StopAll,
	); err != nil {
		// TODO: handle error
	}

	if err := db.Delete(key); err != nil {
		// TODO: handle error
	}
}
