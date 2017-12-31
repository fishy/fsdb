package pool_test

import (
	"testing"

	"github.com/fishy/fsdb/libs/pool"
)

func TestPool(t *testing.T) {
	defer func() {
		if reason := recover(); reason == nil {
			t.Error("Pool.Get() did not call generator on empty pool.")
		}
	}()
	p := pool.NewPool(2)
	gen := func(value int) pool.Generator {
		return func() interface{} {
			return value
		}
	}
	expect := 1
	actual := p.Get(gen(expect))
	if actual != expect {
		t.Errorf("Get expected %v, got %v", expect, actual)
	}
	if size := p.Size(); size != 0 {
		t.Errorf("Pool size expected 0, got: %d", p.Size())
	}
	put := p.Put(expect)
	if !put {
		t.Errorf("Put returned false on non-full pool.")
	}
	expect2 := 2
	put = p.Put(expect2)
	if !put {
		t.Errorf("Put returned false on non-full pool.")
	}
	put = p.Put(3)
	if put {
		t.Errorf("Put returned true on full pool.")
	}
	actual = p.Get(nil)
	if actual != expect {
		t.Errorf("Get expected %v, got %v", expect, actual)
	}
	actual = p.Get(nil)
	if actual != expect2 {
		t.Errorf("Get expected %v, got %v", expect2, actual)
	}
	if size := p.Size(); size != 0 {
		t.Errorf("Pool size expected 0, got: %d", p.Size())
	}
	actual = p.Get(gen(expect))
	if actual != expect {
		t.Errorf("Get expected %v, got %v", expect, actual)
	}
	put = p.Put(expect)
	if !put {
		t.Errorf("Put returned false on non-full pool.")
	}
	actual = p.Get(nil)
	if actual != expect {
		t.Errorf("Get expected %v, got %v", expect, actual)
	}
	actual = p.Get(nil) // Should panic
	t.Errorf("Calling Get(nil) on empty pool should panic, got %v", actual)
}
