package pool_test

import (
	"fmt"

	"github.com/fishy/fsdb/libs/pool"
)

func Example() {
	p := pool.NewPool(1)
	gen := func(value string) pool.Generator {
		return func() interface{} {
			return value
		}
	}

	fmt.Println(p.Size())
	fmt.Println(p.Get(gen("generated data")))
	fmt.Println(p.Size())
	fmt.Println()
	fmt.Println(p.Put("first data"))
	fmt.Println(p.Size())
	fmt.Println(p.Put("second data"))
	fmt.Println(p.Size())
	fmt.Println()
	fmt.Println(p.Get(nil))
	fmt.Println(p.Size())
	fmt.Println()
	fmt.Println(p.Get(gen("generated data")))
	fmt.Println(p.Size())

	// Output:
	// 0
	// generated data
	// 0
	//
	// true
	// 1
	// false
	// 1
	//
	// first data
	// 0
	//
	// generated data
	// 0
}
