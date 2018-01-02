[![GoDoc](https://godoc.org/github.com/fishy/fsdb/fsdb?status.svg)](https://godoc.org/github.com/fishy/fsdb/fsdb)
[![Go Report Card](https://goreportcard.com/badge/github.com/fishy/fsdb)](https://goreportcard.com/report/github.com/fishy/fsdb)

# FSDB

FSDB is a collection of [Go](https://golang.org) libraries providing a key-value
store on your file system.

([Example code on godoc](https://godoc.org/github.com/fishy/fsdb/fsdb/local#example-package))

## Why?

tl;dr: It's for larger (10k+ per entry), less latency sensitive data store.

Most key-value store libraries are optimized for small, in-memory size.
Even for on-disk libraries, they are usually not optimized for larger (10k+)
values.
Also on-disk libraries usually uses write amplify for better performance,
which means they will take more disk space than the actual data stored.
FSDB store the data as-is or use optional gzip compression,
making it a better solution for companies that need to store huge amount of data
and is less sensitive to data latency.

FSDB could also be used as the last layer of your layered key-value stores:
Use in-memory libraries for samllest, most time sensitive data;
Use other on-disk libraries for larger, less time sensitive data;
And use FSDB for largest, least latency sensitive data.

Further more, FSDB provided a hybrid implementation,
which allows you to put some of your data on a remote bucket
(AWS S3, Google Cloud Storage, etc.),
providing an exra layer for larger and higher latency data.

It can also be used to implement mocks of remote libraries.
For example, package [bucket](https://godoc.org/github.com/fishy/fsdb/bucket)
uses local FSDB to implement its
[mock](https://github.com/fishy/fsdb/blob/master/bucket/mock.go) for testing.

## Highlights

FSDB has a minimal overhead,
which means its performance is almost identical to the disk I/O performance.
The local implementation has no locks.
The hybrid implementation only has an optional row lock,
please refer to the
[package documentation](https://godoc.org/github.com/fishy/fsdb/fsdb/hybrid#hdr-Concurrency)
for more details.

The local and hybrid implementation has the same interface,
which means you can use the local implementation first,
and move the the hybrid implementation later as your data grows.

FSDB has no third party dependencies.
All it uses are Go standard libraries and libraries from within this repository.

## Packages Index

### FSDB core packages:

* Package [fsdb](https://godoc.org/github.com/fishy/fsdb/fsdb)
  defines the interface. It does not provide implementations.
* Package [local](https://godoc.org/github.com/fishy/fsdb/fsdb/local)
  provides the local implementation.
* Package [hybrid](https://godoc.org/github.com/fishy/fsdb/fsdb/hybrid)
  provides the hybrid implementation.

### Remote bucket package:

* Package [bucket](https://godoc.org/github.com/fishy/fsdb/bucket)
  defines the bucket interface.
  It does not provide implementations.
  Implementations for [AWS S3](https://godoc.org/github.com/fishy/s3bucket) and
  [Google Cloud Storage](https://godoc.org/github.com/fishy/gcsbucket)
  can be found in external libraries.

### Supportive packages:

* Package [errbatch](https://godoc.org/github.com/fishy/fsdb/libs/errbatch)
  provides
  [ErrBatch](https://godoc.org/github.com/fishy/fsdb/libs/errbatch#ErrBatch),
  which can be used to compile multiple errors into a single error.
* Package [pool](https://godoc.org/github.com/fishy/fsdb/libs/pool)
  provides an implementation of resource pool.
* Package [rowlock](https://godoc.org/github.com/fishy/fsdb/libs/rowlock)
  provides a row lock implementation.
* Package [wrapreader](https://godoc.org/github.com/fishy/fsdb/libs/wrapreader)
  provides a function to
  [Wrap](https://godoc.org/github.com/fishy/fsdb/libs/wrapreader#Wrap) an
  [io.Reader](https://godoc.org/io#Reader) and an
  [io.Closer](https://godoc.org/io#Closer)
  into an [io.ReadCloser](https://godoc.org/io#ReadCloser).

## Test

All packages provide its own tests can be run with `go test`.
If you want to test every package within this repository,
you can use [Bazel](https://bazel.build/) by running `bazel test ...`
under the repository root directory.

There are some tests with sleep calls,
you can skip them by running `go test -short` instead.
Package [local](https://godoc.org/github.com/fishy/fsdb/fsdb/local)
provides a benchmark test can be run with `go test -bench .`.

## License

[BSD License](https://github.com/fishy/fsdb/blob/master/LICENSE).
