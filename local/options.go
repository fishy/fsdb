package local

import (
	"compress/gzip"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"os"
	"strings"

	"github.com/fishy/fsdb/interface"
)

const charsPerLevel = 2

// PathSeparator is the string version of os.PathSeparator.
const PathSeparator = string(os.PathSeparator)

// Default options values.
const (
	DefaultDataDir = "data" + PathSeparator
	DefaultTempDir = "_tmp" + PathSeparator

	DefaultDirLevel = 3

	DefaultUseGzip   = false
	DefaultGzipLevel = gzip.DefaultCompression
	DefaultUseSnappy = false
)

// DefaultHashFunc is the default hash function, which is SHA-512/224.
//
// It's chosen because it gives us relatively shorter hash results,
// thus shorter filenames.
var DefaultHashFunc = sha512.New512_224

// Options defines a read only view of options used by local fsdb.
type Options interface {
	// GetDataDir returns the full path of the root data directory,
	// guaranteed to end with PathSeparator.
	GetDataDir() string

	// GetDataDir returns the full path of the root temporary directory,
	// guaranteed to end with PathSeparator.
	GetTempDir() string

	// GetHashFunc returns the hash function used in keys.
	GetHashFunc() func() hash.Hash

	// GetDirForKey returns the directory to put entry in,
	// guaranteed to end with PathSeparator and guaranteed to be under root data
	// directory.
	GetDirForKey(key fsdb.Key) string

	GetUseGzip() bool
	GetGzipLevel() int

	GetUseSnappy() bool
}

// OptionsBuilder defines a read-write view of options used by local fsdb.
//
// Gzip and Snappy related options are safe to change on an existing FSDB
// system. Changing other options will break the existing FSDB system.
type OptionsBuilder interface {
	Options

	// Build returns the read-only version of options.
	Build() Options

	// SetDataDir sets the relative data directory within the root directory.
	SetDataDir(dir string) OptionsBuilder

	// SetTempDir sets the relative temporary directory within the root directory.
	//
	// It should be on the same mount point as data directory.
	SetTempDir(dir string) OptionsBuilder

	// SetHashFunc sets the hash function used for keys.
	SetHashFunc(f func() hash.Hash) OptionsBuilder

	// SetDirLevel sets the directory level used in filenames.
	// Its purpose is to limit the number of files under the same directory.
	//
	// For example, if directory level was set to 2, hash value "deadbeef" will
	// convert to directory name "de/ad/beef/".
	SetDirLevel(level int) OptionsBuilder

	// SetUseGzip sets whether to use gzip for storage.
	//
	// If gzip is true, this function will also set snappy to false.
	SetUseGzip(gzip bool) OptionsBuilder

	// SetGzipLevel sets the level used in gzip compression.
	SetGzipLevel(level int) OptionsBuilder

	// SetUseSnappy sets whether to use snappy for storage.
	// See https://google.github.io/snappy/ for details.
	//
	// If snappy is true, this function will also set gzip to false.
	SetUseSnappy(snappy bool) OptionsBuilder
}

type options struct {
	root      string
	data      string
	tmp       string
	hashFunc  func() hash.Hash
	dirLevel  int
	useGzip   bool
	gzipLevel int
	useSnappy bool
}

// NewDefaultOptions creates an OptionsBuilder with default options.
func NewDefaultOptions(root string) OptionsBuilder {
	if !strings.HasSuffix(root, PathSeparator) {
		root += PathSeparator
	}
	return &options{
		root:      root,
		data:      DefaultDataDir,
		tmp:       DefaultTempDir,
		hashFunc:  DefaultHashFunc,
		dirLevel:  DefaultDirLevel,
		useGzip:   DefaultUseGzip,
		gzipLevel: DefaultGzipLevel,
		useSnappy: DefaultUseSnappy,
	}
}

func (opts *options) GetDataDir() string {
	return opts.root + opts.data
}

func (opts *options) GetTempDir() string {
	return opts.root + opts.tmp
}

func (opts *options) GetHashFunc() func() hash.Hash {
	return opts.hashFunc
}

func (opts *options) GetDirForKey(key fsdb.Key) string {
	h := opts.GetHashFunc()()
	h.Write(key)
	hashString := hex.EncodeToString(h.Sum([]byte{}))
	path := opts.GetDataDir()
	for i := 0; i < opts.dirLevel; i++ {
		path += hashString[:charsPerLevel]
		path += PathSeparator
		hashString = hashString[charsPerLevel:]
		if len(hashString) <= 0 {
			break
		}
	}
	if len(hashString) > 0 {
		path += hashString
		path += PathSeparator
	}
	return path
}

func (opts *options) GetUseGzip() bool {
	return opts.useGzip
}

func (opts *options) GetGzipLevel() int {
	return opts.gzipLevel
}

func (opts *options) GetUseSnappy() bool {
	return opts.useSnappy
}

func (opts *options) Build() Options {
	return opts
}

func (opts *options) SetDataDir(dir string) OptionsBuilder {
	if !strings.HasSuffix(dir, PathSeparator) {
		dir += PathSeparator
	}
	opts.data = dir
	return opts
}

func (opts *options) SetTempDir(dir string) OptionsBuilder {
	if !strings.HasSuffix(dir, PathSeparator) {
		dir += PathSeparator
	}
	opts.tmp = dir
	return opts
}

func (opts *options) SetHashFunc(f func() hash.Hash) OptionsBuilder {
	opts.hashFunc = f
	return opts
}

func (opts *options) SetDirLevel(level int) OptionsBuilder {
	opts.dirLevel = level
	return opts
}

func (opts *options) SetUseGzip(gzip bool) OptionsBuilder {
	if gzip {
		opts.useSnappy = false
	}
	opts.useGzip = gzip
	return opts
}

func (opts *options) SetGzipLevel(level int) OptionsBuilder {
	opts.gzipLevel = level
	return opts
}

func (opts *options) SetUseSnappy(snappy bool) OptionsBuilder {
	if snappy {
		opts.useGzip = false
	}
	opts.useSnappy = snappy
	return opts
}
