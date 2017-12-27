package remote

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"time"

	"github.com/fishy/fsdb/interface"
)

// Default options values.
const (
	DefaultUploadDelay     time.Duration = time.Minute * 5
	DefaultUploadThreadNum               = 5
)

// DefaultNameFunc is the default name function used.
//
// The format is:
//     fsdb/data/<sha-512/224 of key>.gz
func DefaultNameFunc(key fsdb.Key) string {
	hash := sha512.Sum512_224(key)
	return "fsdb/data/" + hex.EncodeToString(hash[:]) + ".gz"
}

// UploadAll is the skip function that uploads everything to remote bucket.
func UploadAll(key fsdb.Key) bool {
	return false
}

// SkipAll is the skip function that retains everything locally.
func SkipAll(key fsdb.Key) bool {
	return true
}

// DefaultSkipFunc is the default skip function used.
var DefaultSkipFunc = UploadAll

// Options defines a read-only view of options used in remote FSDB.
type Options interface {
	// GetUploadDelay returns the delay between two upload scan loops.
	GetUploadDelay() time.Duration

	// GetUploadThreadNum returns the number of threads used in upload scan loops.
	//
	// The higher the number, the faster the uploads,
	// but it also means heavier I/O.
	GetUploadThreadNum() int

	// GetLogger returns the logger to be used in remote FSDB.
	//
	// If it returns nil, nothing will be logged.
	GetLogger() *log.Logger

	// GetRemoteName returns the name for the data file on remote bucket.
	GetRemoteName(key fsdb.Key) string

	// SkipKey returns true if the key should not be uploaded to remote bucket
	// (retain locally), or false if the key should be uploaded to remote bucket.
	SkipKey(key fsdb.Key) bool

	// It's possible that this function need to read from the remote FSDB,
	// so it's allowed to be changed in read-only Options.
	SetSkipFunc(f func(fsdb.Key) bool)
}

// OptionsBuilder defines a read write view of options used in remote FSDB.
type OptionsBuilder interface {
	Options

	// Build builds the read-only view of the options.
	Build() Options

	// SetUploadDelay sets the delay between two upload scan loops.
	SetUploadDelay(delay time.Duration) OptionsBuilder

	// SetUploadThreadNum sets the number of threads used in upload scan loops.
	SetUploadThreadNum(threads int) OptionsBuilder

	// SetLogger sets the logger used in remote FSDB.
	SetLogger(logger *log.Logger) OptionsBuilder

	// SetRemoteNameFunc sets the function for GetRemoteName.
	SetRemoteNameFunc(f func(fsdb.Key) string) OptionsBuilder
}

type options struct {
	delay    time.Duration
	threads  int
	logger   *log.Logger
	nameFunc func(fsdb.Key) string
	skipFunc func(fsdb.Key) bool
}

// NewDefaultOptions creates the default options.
func NewDefaultOptions() OptionsBuilder {
	return &options{
		delay:    DefaultUploadDelay,
		threads:  DefaultUploadThreadNum,
		logger:   nil,
		nameFunc: DefaultNameFunc,
		skipFunc: DefaultSkipFunc,
	}
}

func (opt *options) GetUploadDelay() time.Duration {
	return opt.delay
}

func (opt *options) GetUploadThreadNum() int {
	return opt.threads
}

func (opt *options) GetLogger() *log.Logger {
	return opt.logger
}

func (opt *options) GetRemoteName(key fsdb.Key) string {
	return opt.nameFunc(key)
}

func (opt *options) SkipKey(key fsdb.Key) bool {
	return opt.skipFunc(key)
}

func (opt *options) Build() Options {
	return opt
}

func (opt *options) SetUploadDelay(delay time.Duration) OptionsBuilder {
	opt.delay = delay
	return opt
}

func (opt *options) SetUploadThreadNum(threads int) OptionsBuilder {
	opt.threads = threads
	return opt
}

func (opt *options) SetLogger(logger *log.Logger) OptionsBuilder {
	opt.logger = logger
	return opt
}

func (opt *options) SetRemoteNameFunc(f func(fsdb.Key) string) OptionsBuilder {
	opt.nameFunc = f
	return opt
}

func (opt *options) SetSkipFunc(f func(fsdb.Key) bool) {
	opt.skipFunc = f
}
