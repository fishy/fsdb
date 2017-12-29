# FSDB

FSDB is a collection of [Go](https://golang.org) libraries providing a key-value
store on your file system.

## Why?

Most key-value store libraries are optimized for small, in-memory size.
Even for on-disk libraries, they are usually not optimized for larger (10k+)
values.
Also on-disk libraries usually uses write amplify for better performance,
which means they will take more disk space than the actual data stored.
FSDB store the data as-is or use optional gzip compression,
making it a better solution for companies that need to store huge amount of data
and is less sensitive to data latency.

FSDB could also be used as the last item of your key-value food chain:
Use in-memory libraries for samllest, most time sensitive data;
Use other on-disk libraries for larger, less time sensitive data;
And use FSDB for largest, least latency sensitive data.

Further more, FSDB provided a hybrid implementation,
which allows you to put some of your data on a remote bucket
(AWS S3, Google Cloud Storage, etc.),
providing an even better price per GB and larger latency.

## Highlights

FSDB has a minimal overhead,
which means its performance is almost identical to the disk I/O performance.
The local implementation has no locks.
The hybrid remote implementation only has an optional row lock,
please refer to the
[package documentation](https://godoc.org/github.com/fishy/fsdb/remote)
for more details.

The local and hybrid implementation has the same interface,
which means you can use the local implementation first,
and move the the hybrid implementation later as your data grows.

FSDB has no third party dependencies.
All it uses are Go standard libraries and libraries from within this repository.

## Packages Index

### FSDB core packages:

* Package [fsdb](https://godoc.org/github.com/fishy/fsdb/interface)
  defines the interface. It does not provide implementations.
* Package [local](https://godoc.org/github.com/fishy/fsdb/local)
  provides the local implementation.
* Package [remote](https://godoc.org/github.com/fishy/fsdb/remote)
  provides the hybrid implementation.

### Remote bucket package:

* Package [bucket](https://godoc.org/github.com/fishy/fsdb/bucket)
  defines the bucket interface.
  It does not provide implementations.
  Also no implementations are provided in this repository,
  but it's easy to implement wrapping S3 or GCS official Go libraries.

### Supportive packages:

* Package [errbatch](https://godoc.org/github.com/fishy/fsdb/errbatch) provides
  [ErrBatch](https://godoc.org/github.com/fishy/fsdb/errbatch#ErrBatch),
  which can be used to compile multiple errors into a single error.
* Package [pool](https://godoc.org/github.com/fishy/fsdb/pool)
  provides an implementation of resource pool.
* Package [rowlock](https://godoc.org/github.com/fishy/fsdb/rowlock)
  provides a row lock implementation.
* Package [wrapreader](https://godoc.org/github.com/fishy/fsdb/wrapreader)
  provides a function to
  [Wrap](https://godoc.org/github.com/fishy/fsdb/wrapreader#Wrap) an
  [io.Reader](https://godoc.org/io#Reader) and an
  [io.Closer](https://godoc.org/io#Closer)
  into an [io.ReadCloser](https://godoc.org/io#ReadCloser).

## License

[BSD License](https://github.com/fishy/fsdb/blob/master/LICENSE).
