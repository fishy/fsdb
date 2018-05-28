package local_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fishy/fsdb"
	"github.com/fishy/fsdb/local"
)

func Example() {
	root, _ := ioutil.TempDir("", "fsdb_")
	defer os.RemoveAll(root)

	db := local.Open(local.NewDefaultOptions(root).SetUseGzip(true))
	key := fsdb.Key("name")
	ctx := context.Background()

	db.Write(ctx, key, strings.NewReader("Anakin Skywalker"))
	reader, err := db.Read(ctx, key)
	if err != nil {
		// TODO: handle error
	}
	name, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		// TODO: handle error
	}
	fmt.Println(string(name))

	db.Write(ctx, key, strings.NewReader("Darth Vader"))
	reader, err = db.Read(ctx, key)
	if err != nil {
		// TODO: handle error
	}
	name, err = ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		// TODO: handle error
	}
	fmt.Println(string(name))

	db.Delete(ctx, key)
	_, err = db.Read(ctx, key)
	if fsdb.IsNoSuchKeyError(err) {
		fmt.Println("Joined force")
	}

	// Output:
	// Anakin Skywalker
	// Darth Vader
	// Joined force
}
