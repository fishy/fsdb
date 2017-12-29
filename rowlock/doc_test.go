package rowlock_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/fishy/fsdb/rowlock"
)

func Example() {
	lock := rowlock.NewRowLock(rowlock.MutexNewLocker)
	key1 := "key1"
	key2 := "key2"
	round := time.Millisecond * 10

	keys := []string{key1, key1, key2, key2}
	sleeps := []time.Duration{
		time.Millisecond * 250,
		time.Millisecond * 200,
		time.Millisecond * 350,
		time.Millisecond * 300,
	}

	var wg sync.WaitGroup
	wg.Add(len(keys))

	for i := range keys {
		go func(key string, sleep time.Duration) {
			started := time.Now()
			defer wg.Done()
			time.Sleep(sleep)
			lock.Lock(key)
			defer lock.Unlock(key)
			elapsed := time.Now().Sub(started).Round(round)
			// The same key with longer sleep will get an elapsed time about
			// 2 * the same key with shorter sleep instead of its own sleep time,
			// because that's when the other goroutine releases the lock.
			fmt.Printf("%s got lock after about %v\n", key, elapsed)
			time.Sleep(sleep)
		}(keys[i], sleeps[i])
	}

	wg.Wait()
	// Output:
	// key1 got lock after about 200ms
	// key2 got lock after about 300ms
	// key1 got lock after about 400ms
	// key2 got lock after about 600ms
}
