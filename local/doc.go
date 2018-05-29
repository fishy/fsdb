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
// A sample result on Debian sid (kernel 4.16) ext4 non-SSD is:
//
//     $ vgo test -bench=.
//     goos: linux
//     goarch: amd64
//     pkg: github.com/fishy/fsdb/local
//     BenchmarkReadWrite/1K/gzip-min/write-2              3000            484932 ns/op
//     BenchmarkReadWrite/1K/gzip-min/read-2              30000             38385 ns/op
//     BenchmarkReadWrite/1K/gzip-default/write-2                  5000            499867 ns/op
//     BenchmarkReadWrite/1K/gzip-default/read-2                  50000             37513 ns/op
//     BenchmarkReadWrite/1K/gzip-max/write-2                      5000            450503 ns/op
//     BenchmarkReadWrite/1K/gzip-max/read-2                      50000             37402 ns/op
//     BenchmarkReadWrite/1K/nocompression/write-2                 5000            492013 ns/op
//     BenchmarkReadWrite/1K/nocompression/read-2                 50000             36236 ns/op
//     BenchmarkReadWrite/10K/nocompression/write-2                5000           1083269 ns/op
//     BenchmarkReadWrite/10K/nocompression/read-2                50000             38601 ns/op
//     BenchmarkReadWrite/10K/gzip-min/write-2                     5000            457293 ns/op
//     BenchmarkReadWrite/10K/gzip-min/read-2                     50000             37044 ns/op
//     BenchmarkReadWrite/10K/gzip-default/write-2                 5000            501657 ns/op
//     BenchmarkReadWrite/10K/gzip-default/read-2                 50000             37068 ns/op
//     BenchmarkReadWrite/10K/gzip-max/write-2                     5000            483487 ns/op
//     BenchmarkReadWrite/10K/gzip-max/read-2                     50000             35943 ns/op
//     BenchmarkReadWrite/1M/gzip-min/write-2                      1000           7456528 ns/op
//     BenchmarkReadWrite/1M/gzip-min/read-2                      50000             37182 ns/op
//     BenchmarkReadWrite/1M/gzip-default/write-2                  1000           8039125 ns/op
//     BenchmarkReadWrite/1M/gzip-default/read-2                  50000             35744 ns/op
//     BenchmarkReadWrite/1M/gzip-max/write-2                      1000          11203349 ns/op
//     BenchmarkReadWrite/1M/gzip-max/read-2                      50000             37702 ns/op
//     BenchmarkReadWrite/1M/nocompression/write-2                 1000           8565408 ns/op
//     BenchmarkReadWrite/1M/nocompression/read-2                 50000             38373 ns/op
//     BenchmarkReadWrite/10M/gzip-default/write-2                  100          72915402 ns/op
//     BenchmarkReadWrite/10M/gzip-default/read-2                 50000             36421 ns/op
//     BenchmarkReadWrite/10M/gzip-max/write-2                      100          79632091 ns/op
//     BenchmarkReadWrite/10M/gzip-max/read-2                     50000             35591 ns/op
//     BenchmarkReadWrite/10M/nocompression/write-2                 100          69915961 ns/op
//     BenchmarkReadWrite/10M/nocompression/read-2                50000             36835 ns/op
//     BenchmarkReadWrite/10M/gzip-min/write-2                      100          66220149 ns/op
//     BenchmarkReadWrite/10M/gzip-min/read-2                     50000             36418 ns/op
//     BenchmarkReadWrite/256M/nocompression/write-2                  2        2380085276 ns/op
//     BenchmarkReadWrite/256M/nocompression/read-2               50000             34536 ns/op
//     BenchmarkReadWrite/256M/gzip-min/write-2                       2        2916555472 ns/op
//     BenchmarkReadWrite/256M/gzip-min/read-2                    50000             34695 ns/op
//     BenchmarkReadWrite/256M/gzip-default/write-2                   2        2347434829 ns/op
//     BenchmarkReadWrite/256M/gzip-default/read-2                50000             35709 ns/op
//     BenchmarkReadWrite/256M/gzip-max/write-2                       2        2311051473 ns/op
//     BenchmarkReadWrite/256M/gzip-max/read-2                    50000             36053 ns/op
//     PASS
//     ok      github.com/fishy/fsdb/local     181.685s
//
// A sample result on macOS 10.13.4 HFS+ hybrid-disk:
//
//     $ vgo test -bench=.
//     goos: darwin
//     goarch: amd64
//     pkg: github.com/fishy/fsdb/local
//     BenchmarkReadWrite/1K/gzip-default/write-4          3000            551077 ns/op
//     BenchmarkReadWrite/1K/gzip-default/read-4          30000             45361 ns/op
//     BenchmarkReadWrite/1K/gzip-max/write-4              2000            579064 ns/op
//     BenchmarkReadWrite/1K/gzip-max/read-4              30000             45244 ns/op
//     BenchmarkReadWrite/1K/nocompression/write-4         3000            589749 ns/op
//     BenchmarkReadWrite/1K/nocompression/read-4         30000             44290 ns/op
//     BenchmarkReadWrite/1K/gzip-min/write-4              3000            601753 ns/op
//     BenchmarkReadWrite/1K/gzip-min/read-4              30000             41715 ns/op
//     BenchmarkReadWrite/10K/gzip-min/write-4             2000            579109 ns/op
//     BenchmarkReadWrite/10K/gzip-min/read-4             30000             41290 ns/op
//     BenchmarkReadWrite/10K/gzip-default/write-4         2000            622533 ns/op
//     BenchmarkReadWrite/10K/gzip-default/read-4         30000             45317 ns/op
//     BenchmarkReadWrite/10K/gzip-max/write-4             3000            592523 ns/op
//     BenchmarkReadWrite/10K/gzip-max/read-4             30000             41447 ns/op
//     BenchmarkReadWrite/10K/nocompression/write-4        2000            640485 ns/op
//     BenchmarkReadWrite/10K/nocompression/read-4        30000             41452 ns/op
//     BenchmarkReadWrite/1M/gzip-min/write-4               300           6990783 ns/op
//     BenchmarkReadWrite/1M/gzip-min/read-4              30000             43776 ns/op
//     BenchmarkReadWrite/1M/gzip-default/write-4           200          10915348 ns/op
//     BenchmarkReadWrite/1M/gzip-default/read-4          30000             42378 ns/op
//     BenchmarkReadWrite/1M/gzip-max/write-4               200          10628149 ns/op
//     BenchmarkReadWrite/1M/gzip-max/read-4              30000             40477 ns/op
//     BenchmarkReadWrite/1M/nocompression/write-4          300           8656265 ns/op
//     BenchmarkReadWrite/1M/nocompression/read-4         30000             44789 ns/op
//     BenchmarkReadWrite/10M/nocompression/write-4          20         114984454 ns/op
//     BenchmarkReadWrite/10M/nocompression/read-4        50000             38741 ns/op
//     BenchmarkReadWrite/10M/gzip-min/write-4               50          71933590 ns/op
//     BenchmarkReadWrite/10M/gzip-min/read-4             50000             39491 ns/op
//     BenchmarkReadWrite/10M/gzip-default/write-4           20          67794399 ns/op
//     BenchmarkReadWrite/10M/gzip-default/read-4         50000             38596 ns/op
//     BenchmarkReadWrite/10M/gzip-max/write-4               30         116057303 ns/op
//     BenchmarkReadWrite/10M/gzip-max/read-4             50000             38180 ns/op
//     BenchmarkReadWrite/256M/gzip-default/write-4           1        1921484298 ns/op
//     BenchmarkReadWrite/256M/gzip-default/read-4        50000             39500 ns/op
//     BenchmarkReadWrite/256M/gzip-max/write-4               1        1657779284 ns/op
//     BenchmarkReadWrite/256M/gzip-max/read-4            50000             40050 ns/op
//     BenchmarkReadWrite/256M/nocompression/write-4                  1        1438590200 ns/op
//     BenchmarkReadWrite/256M/nocompression/read-4               50000             39634 ns/op
//     BenchmarkReadWrite/256M/gzip-min/write-4                       1        1445182668 ns/op
//     BenchmarkReadWrite/256M/gzip-min/read-4                    50000             39145 ns/op
//     PASS
//     ok      github.com/fishy/fsdb/local     96.428s
//
// And a sample result on macOS 10.13.4 HFS+ SSD:
//
//     $ vgo test -bench=.
//     goos: darwin
//     goarch: amd64
//     pkg: github.com/fishy/fsdb/local
//     BenchmarkReadWrite/256M/nocompression/write-8   	       5	 356710471 ns/op
//     BenchmarkReadWrite/256M/nocompression/read-8    	   30000	     43209 ns/op
//     BenchmarkReadWrite/256M/gzip-min/write-8        	       5	 228341933 ns/op
//     BenchmarkReadWrite/256M/gzip-min/read-8         	   30000	     44043 ns/op
//     BenchmarkReadWrite/256M/gzip-default/write-8    	       5	 265429180 ns/op
//     BenchmarkReadWrite/256M/gzip-default/read-8     	   30000	     45661 ns/op
//     BenchmarkReadWrite/256M/gzip-max/write-8        	       5	 280203940 ns/op
//     BenchmarkReadWrite/256M/gzip-max/read-8         	   30000	     44006 ns/op
//     BenchmarkReadWrite/1K/nocompression/write-8     	    2000	    824098 ns/op
//     BenchmarkReadWrite/1K/nocompression/read-8      	   30000	     45584 ns/op
//     BenchmarkReadWrite/1K/gzip-min/write-8          	    2000	    766742 ns/op
//     BenchmarkReadWrite/1K/gzip-min/read-8           	   20000	     85293 ns/op
//     BenchmarkReadWrite/1K/gzip-default/write-8      	    2000	    858178 ns/op
//     BenchmarkReadWrite/1K/gzip-default/read-8       	   30000	     59477 ns/op
//     BenchmarkReadWrite/1K/gzip-max/write-8          	    2000	    839374 ns/op
//     BenchmarkReadWrite/1K/gzip-max/read-8           	   30000	     46870 ns/op
//     BenchmarkReadWrite/10K/nocompression/write-8    	    2000	    805031 ns/op
//     BenchmarkReadWrite/10K/nocompression/read-8     	   20000	     51670 ns/op
//     BenchmarkReadWrite/10K/gzip-min/write-8         	    2000	    929401 ns/op
//     BenchmarkReadWrite/10K/gzip-min/read-8          	   20000	     80976 ns/op
//     BenchmarkReadWrite/10K/gzip-default/write-8     	    2000	    818654 ns/op
//     BenchmarkReadWrite/10K/gzip-default/read-8      	   30000	     44932 ns/op
//     BenchmarkReadWrite/10K/gzip-max/write-8         	    2000	    752227 ns/op
//     BenchmarkReadWrite/10K/gzip-max/read-8          	   30000	     47205 ns/op
//     BenchmarkReadWrite/1M/gzip-min/write-8          	    1000	   1371292 ns/op
//     BenchmarkReadWrite/1M/gzip-min/read-8           	   10000	    101911 ns/op
//     BenchmarkReadWrite/1M/gzip-default/write-8      	    1000	   1347627 ns/op
//     BenchmarkReadWrite/1M/gzip-default/read-8       	   20000	     54486 ns/op
//     BenchmarkReadWrite/1M/gzip-max/write-8          	    1000	   1408124 ns/op
//     BenchmarkReadWrite/1M/gzip-max/read-8           	   20000	     51155 ns/op
//     BenchmarkReadWrite/1M/nocompression/write-8     	    1000	   1238631 ns/op
//     BenchmarkReadWrite/1M/nocompression/read-8      	   30000	     45901 ns/op
//     BenchmarkReadWrite/10M/nocompression/write-8    	     100	  10761830 ns/op
//     BenchmarkReadWrite/10M/nocompression/read-8     	   30000	     42879 ns/op
//     BenchmarkReadWrite/10M/gzip-min/write-8         	     100	  11185495 ns/op
//     BenchmarkReadWrite/10M/gzip-min/read-8          	   30000	     43027 ns/op
//     BenchmarkReadWrite/10M/gzip-default/write-8     	     100	  11036515 ns/op
//     BenchmarkReadWrite/10M/gzip-default/read-8      	   30000	     43005 ns/op
//     BenchmarkReadWrite/10M/gzip-max/write-8         	     100	  11564331 ns/op
//     BenchmarkReadWrite/10M/gzip-max/read-8          	   30000	     43158 ns/op
//     PASS
//     ok  	github.com/fishy/fsdb/local	88.676s
//
// Other Notes
//
// Remember to set appropriate number of file number limit on your filesystem.
package local
