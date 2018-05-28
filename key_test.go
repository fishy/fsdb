package fsdb_test

import (
	"testing"

	"github.com/fishy/fsdb"
)

func TestString(t *testing.T) {
	expect := "foobar"
	key := fsdb.Key(expect)
	actual := key.String()
	if expect != actual {
		t.Errorf("(%v).String() expected %q, got %q", key, expect, actual)
	}

	key = fsdb.Key{0xff, 0xfe, 0xfd}
	expect = "[255 254 253]"
	actual = key.String()
	if expect != actual {
		t.Errorf("(%v).String() expected %q, got %q", key, expect, actual)
	}
}

func TestEquals(t *testing.T) {
	key1 := fsdb.Key("foobar")
	key2 := fsdb.Key("foobar")
	if !key1.Equals(key2) {
		t.Errorf("%v should equal to %+v", key1, key2)
	}
	if !key2.Equals(key1) {
		t.Errorf("%v should equal to %+v", key2, key1)
	}

	key2 = fsdb.Key{0xff, 0xfe, 0xfd}
	if key1.Equals(key2) {
		t.Errorf("%v should not equal to %+v", key1, key2)
	}
	if key2.Equals(key1) {
		t.Errorf("%v should not equal to %+v", key2, key1)
	}
}
