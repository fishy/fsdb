// Package hybrid provides a hybrid FSDB implementation.
//
// A hybrid FSDB is backed by a local FSDB and a remote bucket.
// All data are written locally first, then a background thread will upload them
// to the remote bucket and delete the local data.
// Read operations will check local FSDB first,
// and fetch from bucket if it does not present locally.
// When remote read happens,
// the data will be saved locally until the next upload loop.
//
// Data stored on the remote bucket will be gzipped using best compression
// level.
//
// Concurrency
//
// If you turn off the optional row lock (default is on),
// there are two possible cases we might lose date due to race conditions,
// but they are very unlikely.
//
// The first case is remote read. The read process is:
//     1. Check local FSDB.
//     2. Read fully from remote bucket.
//     3. Check local FSDB again to prevent using stale remote data to overwrite local data.
//     4. If there's still no local data in Step 3, write remote data locally.
//     5. Return local data.
// If another overwrite happens between Step 3 and 4,
// then it might be overwritten by stale remote data.
//
// The other case is during upload. The upload process for each key is:
//     1. Read local data, calculate crc32c.
//     2. Gzip local data, upload to remote bucket.
//     3. Calculate local data crc32c again.
//     4. If the crc32c from Step 1 and Step 3 matches, delete local data.
// If another overwrite happens between Step 3 and 4,
// then it might be deleted on Step 4 so we only have stale data in the system.
//
// Turning on the optional row lock will make sure the discussed data loss
// scenarios won't happen, but it also degrade the performance slightly.
// The lock is only used partially inside the operations
// (whole local write operation, remote read from Step 3, upload from Step 3).
//
// There are no other locks used in the code,
// except a few atomic numbers in upload loop for logging purpose.
package hybrid
