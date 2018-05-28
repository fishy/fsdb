package fsdb_test

import (
	"errors"
	"testing"

	"github.com/fishy/fsdb"
)

func TestError(t *testing.T) {
	err := &fsdb.NoSuchKeyError{
		Key: fsdb.Key("foobar"),
	}
	expect := "no such key: \"foobar\""
	actual := err.Error()
	if expect != actual {
		t.Errorf("(%q).Error() expected %q, got %q", err, expect, actual)
	}
}

func TestTypeCheck(t *testing.T) {
	var err error

	err = &fsdb.NoSuchKeyError{
		Key: fsdb.Key("foobar"),
	}
	if !fsdb.IsNoSuchKeyError(err) {
		t.Errorf("%q should be an instance of NoSuchKeyError", err)
	}

	err = errors.New("foobar")
	if fsdb.IsNoSuchKeyError(err) {
		t.Errorf("%q should not be an instance of NoSuchKeyError", err)
	}

	if fsdb.IsNoSuchKeyError(nil) {
		t.Errorf("nil should not be an instance of NoSuchKeyError")
	}
}
