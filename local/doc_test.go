package local_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fishy/fsdb/interface"
	"github.com/fishy/fsdb/local"
)

func Example() {
	db := local.Open(local.NewDefaultOptions("_tmp/fsdb_root").SetUseGzip(true))
	key := fsdb.Key("name")

	db.Write(key, strings.NewReader("Anakin Skywalker"))
	reader, err := db.Read(key)
	if err != nil {
		// TODO: handle error
	}
	name, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		// TODO: handle error
	}
	fmt.Println(string(name))

	db.Write(key, strings.NewReader("Darth Vader"))
	reader, err = db.Read(key)
	if err != nil {
		// TODO: handle error
	}
	name, err = ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		// TODO: handle error
	}
	fmt.Println(string(name))

	db.Delete(key)
	_, err = db.Read(key)
	if fsdb.IsNoSuchKeyError(err) {
		fmt.Println("Joined force")
	}

	// Cleanup for go test, no need in prod service.
	os.RemoveAll("_tmp")
	// Output:
	// Anakin Skywalker
	// Darth Vader
	// Joined force
}
