package fsdb

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// Key is the key type of an FSDB.
type Key []byte

func (key Key) String() string {
	if utf8.Valid(key) {
		return string(key)
	}
	return fmt.Sprintf("%v", []byte(key))
}

// Equals compares the key to another key.
func (key Key) Equals(other Key) bool {
	return bytes.Equal(key, other)
}
