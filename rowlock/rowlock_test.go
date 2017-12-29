package rowlock_test

import (
	"sync"
	"testing"
	"time"

	"github.com/fishy/fsdb/rowlock"
)

func TestRowLock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	lock := rowlock.NewRowLock(rowlock.MutexNewLocker)
	key1 := "key1"
	key2 := "key2"

	short := time.Millisecond * 10
	long := time.Millisecond * 100
	longer := time.Millisecond * 150

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		lock.Lock(key1)
		defer lock.Unlock(key1)
		time.Sleep(long)
	}()

	go func() {
		defer wg.Done()
		lock.Lock(key2)
		defer lock.Unlock(key2)
		time.Sleep(longer)
	}()

	go func() {
		defer wg.Done()
		started := time.Now()
		time.Sleep(short)
		lock.Lock(key1)
		defer lock.Unlock(key1)
		elapsed := time.Now().Sub(started)
		t.Logf("elapsed time: %v", elapsed)
		if elapsed < long || elapsed > longer {
			t.Errorf(
				"lock wait time should be between %v and %v, actual %v",
				long,
				longer,
				elapsed,
			)
		}
	}()

	wg.Wait()
}
