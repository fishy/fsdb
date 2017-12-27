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
// This implementation supports the following compression options:
//     - no compression
//     - gzip (configurable levels)
//     - snappy (https://google.github.io/snappy/)
//
// If you changed the compression option on a non-empty local fsdb,
// the old data is still readable and the new data will be stored per new
// compression option.
//
// Run
//     go test -bench .
// will show you the read and write benchmark results of different compression
// options of the filesystem on
//     $PWD/_fsdb_tmp/
// for different typical sizes of random binary data.
//
// You should choose your compression options based on your benchmark result,
// typical data size and estimated read/write operation ratio.
//
// A sample result on Debian sid ext4 non-SSD is:
//     $ go test -bench .
//     goos: linux
//     goarch: amd64
//     pkg: github.com/fishy/fsdb/local
//     BenchmarkReadWrite/256M/snappy/write-2         	       2	2644709778 ns/op
//     BenchmarkReadWrite/256M/snappy/read-2          	   20000	     71461 ns/op
//     BenchmarkReadWrite/256M/gzip-min/write-2       	       2	2524391920 ns/op
//     BenchmarkReadWrite/256M/gzip-min/read-2        	   50000	     28098 ns/op
//     BenchmarkReadWrite/256M/gzip-default/write-2   	       2	2933841783 ns/op
//     BenchmarkReadWrite/256M/gzip-default/read-2    	   50000	     27950 ns/op
//     BenchmarkReadWrite/256M/gzip-max/write-2       	       2	2728275678 ns/op
//     BenchmarkReadWrite/256M/gzip-max/read-2        	   50000	     27749 ns/op
//     BenchmarkReadWrite/256M/nocompression/write-2  	       2	2694980754 ns/op
//     BenchmarkReadWrite/256M/nocompression/read-2   	   50000	     28001 ns/op
//     BenchmarkReadWrite/1K/gzip-default/write-2     	    5000	    277333 ns/op
//     BenchmarkReadWrite/1K/gzip-default/read-2      	   50000	     29347 ns/op
//     BenchmarkReadWrite/1K/gzip-max/write-2         	    5000	    279393 ns/op
//     BenchmarkReadWrite/1K/gzip-max/read-2          	   50000	     29365 ns/op
//     BenchmarkReadWrite/1K/nocompression/write-2    	    5000	    300942 ns/op
//     BenchmarkReadWrite/1K/nocompression/read-2     	   50000	     29854 ns/op
//     BenchmarkReadWrite/1K/snappy/write-2           	    3000	    371906 ns/op
//     BenchmarkReadWrite/1K/snappy/read-2            	   20000	     81126 ns/op
//     BenchmarkReadWrite/1K/gzip-min/write-2         	    5000	    278225 ns/op
//     BenchmarkReadWrite/1K/gzip-min/read-2          	   50000	     29681 ns/op
//     BenchmarkReadWrite/10K/gzip-min/write-2        	    5000	    299979 ns/op
//     BenchmarkReadWrite/10K/gzip-min/read-2         	   50000	     28481 ns/op
//     BenchmarkReadWrite/10K/gzip-default/write-2    	    5000	    270725 ns/op
//     BenchmarkReadWrite/10K/gzip-default/read-2     	   50000	     28427 ns/op
//     BenchmarkReadWrite/10K/gzip-max/write-2        	    5000	    289610 ns/op
//     BenchmarkReadWrite/10K/gzip-max/read-2         	   50000	     28708 ns/op
//     BenchmarkReadWrite/10K/nocompression/write-2   	    5000	    290857 ns/op
//     BenchmarkReadWrite/10K/nocompression/read-2    	   50000	     29612 ns/op
//     BenchmarkReadWrite/10K/snappy/write-2          	    5000	    408231 ns/op
//     BenchmarkReadWrite/10K/snappy/read-2           	   20000	     80573 ns/op
//     BenchmarkReadWrite/1M/gzip-max/write-2         	    1000	   1498500 ns/op
//     BenchmarkReadWrite/1M/gzip-max/read-2          	   50000	     29521 ns/op
//     BenchmarkReadWrite/1M/nocompression/write-2    	    1000	   1540354 ns/op
//     BenchmarkReadWrite/1M/nocompression/read-2     	   50000	     29962 ns/op
//     BenchmarkReadWrite/1M/snappy/write-2           	    1000	   2226409 ns/op
//     BenchmarkReadWrite/1M/snappy/read-2            	   20000	     81671 ns/op
//     BenchmarkReadWrite/1M/gzip-min/write-2         	    1000	   1615656 ns/op
//     BenchmarkReadWrite/1M/gzip-min/read-2          	   50000	     30307 ns/op
//     BenchmarkReadWrite/1M/gzip-default/write-2     	    1000	   1548937 ns/op
//     BenchmarkReadWrite/1M/gzip-default/read-2      	   50000	     29673 ns/op
//     BenchmarkReadWrite/10M/snappy/write-2          	     100	  17206851 ns/op
//     BenchmarkReadWrite/10M/snappy/read-2           	   20000	     75328 ns/op
//     BenchmarkReadWrite/10M/gzip-min/write-2        	     100	  12804069 ns/op
//     BenchmarkReadWrite/10M/gzip-min/read-2         	   50000	     28253 ns/op
//     BenchmarkReadWrite/10M/gzip-default/write-2    	     100	  12490179 ns/op
//     BenchmarkReadWrite/10M/gzip-default/read-2     	   50000	     28300 ns/op
//     BenchmarkReadWrite/10M/gzip-max/write-2        	     100	  12511874 ns/op
//     BenchmarkReadWrite/10M/gzip-max/read-2         	   50000	     28740 ns/op
//     BenchmarkReadWrite/10M/nocompression/write-2   	     100	  12571074 ns/op
//     BenchmarkReadWrite/10M/nocompression/read-2    	   50000	     27684 ns/op
//     PASS
//     ok  	github.com/fishy/fsdb/local	115.508s
//
// Other Notes
//
// Remember to set appropriate number of files limit on your filesystem.
package local
