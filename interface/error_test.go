package fsdb_test

import (
	"errors"
	"testing"

	"github.com/fishy/fsdb/interface"
)

func TestError(t *testing.T) {
	err := fsdb.NewErrNoSuchKey(fsdb.Key("foobar"))
	expect := "no such key: \"foobar\""
	actual := err.Error()
	if expect != actual {
		t.Errorf("(%q).Error() expected %q, got %q", err, expect, actual)
	}
}

func TestTypeCheck(t *testing.T) {
	var err error

	err = fsdb.NewErrNoSuchKey(fsdb.Key("foobar"))
	if !fsdb.IsErrNoSuchKey(err) {
		t.Errorf("%q should be an instance of ErrNoSuchKey", err)
	}

	err = errors.New("foobar")
	if fsdb.IsErrNoSuchKey(err) {
		t.Errorf("%q should not be an instance of ErrNoSuchKey", err)
	}
}
