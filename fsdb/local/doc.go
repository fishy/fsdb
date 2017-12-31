// Package local provides an implementation of key-value store on your
// filesystem.
//
// It implements fsdb.Local interface.
//
// Layout
//
// With default options (SHA-512/224 hash, 3 directory levels),
// a key of "key" will be hashed into
//     6cb1b0e50d74419e2244eaa7328235f71b48c7e1c33b23f6f9517d14
// and the files will be stored under:
//     <fsdb-root>/
//       data/
//         6c/
//           b1/
//             b0/
//               e50d74419e2244eaa7328235f71b48c7e1c33b23f6f9517d14/
//                 key     // Key file
//                 data    // Data file if no compression
//                 data.gz // Data file if gzip enabled
//
// There could also be temporary files for unfinished write operations under
//     <fsdb-root>/_tmp/fsdb_<tmpdir>/
//
// Both hash function and directory levels are configurable.
//
// Atomicity
//
// There's no extra locks in the implementation.
// The atomicity relies on the atomicity guaranteed by your filesystem on
// operations like move (rename), delete, open, etc.
//
// Read Before Overwriting Finishes on the Same Key
//
// If you issue a read operation before an overwrite operation (write operation
// on an existing key) on the same key finishes (returns),
// based on your timing you could get either the previous data,
// or new data.
// You will never read corrupt/incomplete data when fsdb returns nil error.
//
// In details, the write operation sequence is:
//
//     1. Check for key collision
//     2. Write key-value data onto temporary directory
//     3. Move new key-value data from temporary directory to actual directory
//     4. Delete old value(s), if any
//
// Read operations issued before Step 3 will get the old data.
// Read operations issued after Step 3 will get the new data.
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
// corresponding write benchmark test only ran for a few times
// (large size sample, like the 256M benchmark test results below),
// as those read benchmark tests will read from the same file over and over
// again, and most filesystem will optimize for such use case.
//
// You should choose your compression options based on your benchmark result,
// typical data size and estimated read/write operation ratio.
//
// A sample result on Debian sid (kernel 4.14) ext4 non-SSD is:
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
//     ok  	github.com/fishy/fsdb/fsdb/local	192.207s
//
// And a sample result on macOS 10.12.6 HFS+ SSD:
//     $ go test -bench .
//     goos: darwin
//     goarch: amd64
//     pkg: github.com/fishy/fsdb/fsdb/local
//     BenchmarkReadWrite/1K/nocompression/write-8         2000            627496 ns/op
//     BenchmarkReadWrite/1K/nocompression/read-8         50000             26234 ns/op
//     BenchmarkReadWrite/1K/gzip-min/write-8              3000            585881 ns/op
//     BenchmarkReadWrite/1K/gzip-min/read-8              50000             26138 ns/op
//     BenchmarkReadWrite/1K/gzip-default/write-8          3000            647935 ns/op
//     BenchmarkReadWrite/1K/gzip-default/read-8          50000             29446 ns/op
//     BenchmarkReadWrite/1K/gzip-max/write-8              3000            593911 ns/op
//     BenchmarkReadWrite/1K/gzip-max/read-8              50000             27001 ns/op
//     BenchmarkReadWrite/10K/nocompression/write-8        3000            630647 ns/op
//     BenchmarkReadWrite/10K/nocompression/read-8        50000             31496 ns/op
//     BenchmarkReadWrite/10K/gzip-min/write-8             2000            597253 ns/op
//     BenchmarkReadWrite/10K/gzip-min/read-8             50000             30775 ns/op
//     BenchmarkReadWrite/10K/gzip-default/write-8         3000            619333 ns/op
//     BenchmarkReadWrite/10K/gzip-default/read-8         50000             29823 ns/op
//     BenchmarkReadWrite/10K/gzip-max/write-8             3000            563971 ns/op
//     BenchmarkReadWrite/10K/gzip-max/read-8             50000             29140 ns/op
//     BenchmarkReadWrite/1M/nocompression/write-8         2000           1106447 ns/op
//     BenchmarkReadWrite/1M/nocompression/read-8         50000             27095 ns/op
//     BenchmarkReadWrite/1M/gzip-min/write-8              2000           1116115 ns/op
//     BenchmarkReadWrite/1M/gzip-min/read-8              50000             27201 ns/op
//     BenchmarkReadWrite/1M/gzip-default/write-8          2000           1140712 ns/op
//     BenchmarkReadWrite/1M/gzip-default/read-8          50000             26646 ns/op
//     BenchmarkReadWrite/1M/gzip-max/write-8              2000            983144 ns/op
//     BenchmarkReadWrite/1M/gzip-max/read-8              50000             26480 ns/op
//     BenchmarkReadWrite/10M/nocompression/write-8         100          10694977 ns/op
//     BenchmarkReadWrite/10M/nocompression/read-8        50000             24534 ns/op
//     BenchmarkReadWrite/10M/gzip-min/write-8              100          10443349 ns/op
//     BenchmarkReadWrite/10M/gzip-min/read-8            100000             23875 ns/op
//     BenchmarkReadWrite/10M/gzip-default/write-8          200          10210270 ns/op
//     BenchmarkReadWrite/10M/gzip-default/read-8         50000             24585 ns/op
//     BenchmarkReadWrite/10M/gzip-max/write-8              100          11050222 ns/op
//     BenchmarkReadWrite/10M/gzip-max/read-8            100000             23358 ns/op
//     BenchmarkReadWrite/256M/nocompression/write-8          5         318933511 ns/op
//     BenchmarkReadWrite/256M/nocompression/read-8      100000             21785 ns/op
//     BenchmarkReadWrite/256M/gzip-min/write-8               5         337559416 ns/op
//     BenchmarkReadWrite/256M/gzip-min/read-8           100000             21841 ns/op
//     BenchmarkReadWrite/256M/gzip-default/write-8           5         304396763 ns/op
//     BenchmarkReadWrite/256M/gzip-default/read-8       100000             21917 ns/op
//     BenchmarkReadWrite/256M/gzip-max/write-8               5         307852001 ns/op
//     BenchmarkReadWrite/256M/gzip-max/read-8           100000             21520 ns/op
//     PASS
//     ok      github.com/fishy/fsdb/fsdb/local     99.760s
//
// Other Notes
//
// Remember to set appropriate number of file number limit on your filesystem.
package local
