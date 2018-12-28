[![GoDoc](https://godoc.org/github.com/fishy/fsdb?status.svg)](https://godoc.org/github.com/fishy/fsdb)
[![Go Report Card](https://goreportcard.com/badge/github.com/fishy/fsdb)](https://goreportcard.com/report/github.com/fishy/fsdb)

# FSDB

FSDB is a collection of [Go](https://golang.org) libraries providing a key-value
store on your file system.

([Example code on godoc](https://godoc.org/github.com/fishy/fsdb/local#example-package))

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
[package documentation](https://godoc.org/github.com/fishy/fsdb/hybrid#hdr-Concurrency)
for more details.

The local and hybrid implementation has the same interface,
which means you can use the local implementation first,
and move the the hybrid implementation later as your data grows.

## Packages Index

* Package [fsdb](https://godoc.org/github.com/fishy/fsdb)
  defines the interface. It does not provide implementations.
* Package [local](https://godoc.org/github.com/fishy/fsdb/local)
  provides the local implementation.
* Package [hybrid](https://godoc.org/github.com/fishy/fsdb/hybrid)
  provides the hybrid implementation.
* Package [bucket](https://godoc.org/github.com/fishy/fsdb/bucket)
  defines the bucket interface.
  It does not provide implementations.
  Implementations for [AWS S3](https://godoc.org/github.com/fishy/s3bucket) and
  [Google Cloud Storage](https://godoc.org/github.com/fishy/gcsbucket)
  can be found in external libraries.
  There's also an
  [implementation](https://godoc.org/github.com/fishy/blobbucket)
  based on
  [Go-Cloud](https://github.com/google/go-cloud)
  [Blob](https://godoc.org/gocloud.dev/blob)
	interface so you can use any Go-Cloud Blob implementation in hybrid FSDB.

## Test

All packages provide its own tests can be run with `go test`.
If you want to test every package within this repository,
on Go 1.11+ you can use `go test ./...`.
If you are on an older version of Go,
you can use [vgo](https://github.com/golang/vgo/) by running `vgo test all`
under the repository root directory.

There are some tests with sleep calls,
you can skip them by running `go test -short` instead.
Package [local](https://godoc.org/github.com/fishy/fsdb/local)
provides a benchmark test can be run with `go test -bench .`.

## License

[BSD License](https://github.com/fishy/fsdb/blob/master/LICENSE).
