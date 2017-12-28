// Package local provides an implementation of key-value store on your
// filesystem.
//
// It implements fsdb.Local interface.
//
// Atomicity
//
// There's no extra locks in the implementation.
// The atomicity relies on the atomicity guarantee of your filesystem on
// operations like move (rename), delete, open, etc.
//
// Read Before Overwriting Finishes on the Same Key
//
// If you issue a read operation before a overwrite operation (write operation
// on an existing key) on the same key finishes (returns),
// based on your timing you could get either the previous data, or empty data
// (less likely).
// You will never read corrupt/incomplete data when fsdb returns nil error.
//
// In details, the write operation sequence is:
//
//     1. Check for key collision
//     2. Write key-value data onto temporary directory
//     3. Delete old value(s), if any
//     4. Move new key-value data from temporary directory to actual directory
//
// Read operations issued before Step 3 will get the old data.
// Read operations issued while Step 3 and 4 are running will get empty data.
//
// Two Write Operations on the Same Key
//
// If you issue a write operation before another write operation on the same key
// finishes, the one that finishes first will be overwritten by the other.
//
// Compression
//
// This implementation supports optional gzip compression with configurable
// compression levels.
//
// If you changed the compression option on a non-empty local fsdb,
// the old data is still readable and the new data will be stored per new
// compression option.
//
// Run
//     go test -bench .
// will show you the read and write benchmark results of different compression
// options of the filesystem under current directory for different typical sizes
// of random binary data.
// Please note that the write time includes the time used to generate the random
// binary data.
// Also note that the read time could be much smaller than reality when the
// corresponding write benchmark test only ran for a few times (large size),
// as those read benchmark tests will read from the same file over and over
// again, and most filesystem will optimize for such use case.
//
// You should choose your compression options based on your benchmark result,
// typical data size and estimated read/write operation ratio.
//
// A sample result on Debian sid ext4 non-SSD is:
//     $ go test -bench .
//     goos: linux
//     goarch: amd64
//     BenchmarkReadWrite/10K/gzip-default/write-2   	    5000	    424289 ns/op
//     BenchmarkReadWrite/10K/gzip-default/read-2    	   50000	     31666 ns/op
//     BenchmarkReadWrite/10K/gzip-max/write-2       	    5000	    406294 ns/op
//     BenchmarkReadWrite/10K/gzip-max/read-2        	   50000	     32462 ns/op
//     BenchmarkReadWrite/10K/nocompression/write-2  	    5000	    466215 ns/op
//     BenchmarkReadWrite/10K/nocompression/read-2   	   50000	     31562 ns/op
//     BenchmarkReadWrite/10K/gzip-min/write-2       	    5000	    452478 ns/op
//     BenchmarkReadWrite/10K/gzip-min/read-2        	   50000	     30205 ns/op
//     BenchmarkReadWrite/1M/gzip-max/write-2        	    2000	  12568642 ns/op
//     BenchmarkReadWrite/1M/gzip-max/read-2         	    1000	   1101386 ns/op
//     BenchmarkReadWrite/1M/nocompression/write-2   	    1000	  11662001 ns/op
//     BenchmarkReadWrite/1M/nocompression/read-2    	   50000	     31565 ns/op
//     BenchmarkReadWrite/1M/gzip-min/write-2        	    1000	  10113306 ns/op
//     BenchmarkReadWrite/1M/gzip-min/read-2         	   50000	     30968 ns/op
//     BenchmarkReadWrite/1M/gzip-default/write-2    	    2000	  12757153 ns/op
//     BenchmarkReadWrite/1M/gzip-default/read-2     	    2000	    800529 ns/op
//     BenchmarkReadWrite/10M/gzip-default/write-2   	     100	  82405077 ns/op
//     BenchmarkReadWrite/10M/gzip-default/read-2    	   50000	     29945 ns/op
//     BenchmarkReadWrite/10M/gzip-max/write-2       	     100	  82221152 ns/op
//     BenchmarkReadWrite/10M/gzip-max/read-2        	   50000	     30059 ns/op
//     BenchmarkReadWrite/10M/nocompression/write-2  	     100	  74676057 ns/op
//     BenchmarkReadWrite/10M/nocompression/read-2   	   50000	     29943 ns/op
//     BenchmarkReadWrite/10M/gzip-min/write-2       	     100	  77441153 ns/op
//     BenchmarkReadWrite/10M/gzip-min/read-2        	   50000	     29468 ns/op
//     BenchmarkReadWrite/256M/nocompression/write-2 	       1	1093422819 ns/op
//     BenchmarkReadWrite/256M/nocompression/read-2  	   50000	     28690 ns/op
//     BenchmarkReadWrite/256M/gzip-min/write-2      	       1	1006354251 ns/op
//     BenchmarkReadWrite/256M/gzip-min/read-2       	   50000	     28624 ns/op
//     BenchmarkReadWrite/256M/gzip-default/write-2  	       1	1081434172 ns/op
//     BenchmarkReadWrite/256M/gzip-default/read-2   	   50000	     28835 ns/op
//     BenchmarkReadWrite/256M/gzip-max/write-2      	       1	1098102178 ns/op
//     BenchmarkReadWrite/256M/gzip-max/read-2       	   50000	     28541 ns/op
//     BenchmarkReadWrite/1K/gzip-min/write-2        	    5000	    562206 ns/op
//     BenchmarkReadWrite/1K/gzip-min/read-2         	   50000	     30493 ns/op
//     BenchmarkReadWrite/1K/gzip-default/write-2    	    5000	    664398 ns/op
//     BenchmarkReadWrite/1K/gzip-default/read-2     	   50000	     31062 ns/op
//     BenchmarkReadWrite/1K/gzip-max/write-2        	    5000	    649059 ns/op
//     BenchmarkReadWrite/1K/gzip-max/read-2         	   50000	     30319 ns/op
//     BenchmarkReadWrite/1K/nocompression/write-2   	    5000	    657600 ns/op
//     BenchmarkReadWrite/1K/nocompression/read-2    	   50000	     30951 ns/op
//     PASS
//     ok  	github.com/fishy/fsdb/local	192.207s
//
// Other Notes
//
// Remember to set appropriate number of file number limit on your filesystem.
package local
