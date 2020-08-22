[![PkgGoDev](https://pkg.go.dev/badge/github.com/fishy/fsdb)](https://pkg.go.dev/github.com/fishy/fsdb)
[![Go Report Card](https://goreportcard.com/badge/github.com/fishy/fsdb)](https://goreportcard.com/report/github.com/fishy/fsdb)

# FSDB

FSDB is a collection of [Go](https://golang.org) libraries providing a key-value
store on your file system.

([Example code](https://pkg.go.dev/github.com/fishy/fsdb/local?tab=doc#example-package))

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
For example, package [bucket](https://pkg.go.dev/github.com/fishy/fsdb/bucket)
uses local FSDB to implement its
[mock](https://github.com/fishy/fsdb/blob/master/bucket/mock.go) for testing.

## Highlights

FSDB has a minimal overhead,
which means its performance is almost identical to the disk I/O performance.
The local implementation has no locks.
The hybrid implementation only has an optional row lock,
please refer to the
[package documentation](https://pkg.go.dev/github.com/fishy/fsdb/hybrid?tab=doc#hdr-Concurrency)
for more details.

The local and hybrid implementation has the same interface,
which means you can use the local implementation first,
and move the the hybrid implementation later as your data grows.

## Packages Index

* Package [fsdb](https://pkg.go.dev/github.com/fishy/fsdb)
  defines the interface. It does not provide implementations.
* Package [local](https://pkg.go.dev/github.com/fishy/fsdb/local)
  provides the local implementation.
* Package [hybrid](https://pkg.go.dev/github.com/fishy/fsdb/hybrid)
  provides the hybrid implementation.
* Package [bucket](https://pkg.go.dev/github.com/fishy/fsdb/bucket)
  defines the bucket interface.
  It does not provide implementations.
  Implementations for [AWS S3](https://pkg.go.dev/github.com/fishy/s3bucket) and
  [Google Cloud Storage](https://pkg.go.dev/github.com/fishy/gcsbucket)
  can be found in external libraries.
  There's also an
  [implementation](https://pkg.go.dev/github.com/fishy/blobbucket)
  based on
  [Go-Cloud](https://github.com/google/go-cloud)
  [Blob](https://pkg.go.dev/gocloud.dev/blob)
	interface so you can use any Go-Cloud Blob implementation in hybrid FSDB.

## Test

All packages provide its own tests can be run with `go test`.
If you want to test every package within this repository,
on Go 1.11+ you can use `go test ./...`.

There are some tests with sleep calls,
you can skip them by running `go test -short` instead.
Package [local](https://pkg.go.dev/github.com/fishy/fsdb/local)
provides a benchmark test can be run with `go test -bench .`.

## License

[BSD License](https://github.com/fishy/fsdb/blob/master/LICENSE).
