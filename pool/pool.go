package pool

import (
	"sync"
)

// Generator is the function to generate a new resource when getting from an empty
// pool.
type Generator func() interface{}

type node struct {
	resource interface{}
	next     *node
}

// Pool is a resource pool.
//
// It's implemented as a linked array.
//
// In most cases, there's no need to prefill the pool.
type Pool struct {
	size    int
	maxSize int
	head    *node
	tail    *node
	locker  sync.Locker
}

// NewPool creates a new pool.
//
// maxSize can be used to limit the number of resources stored in the pool.
// if maxSize <= 0, the size of the pool is unlimited.
func NewPool(maxSize int) *Pool {
	return &Pool{
		maxSize: maxSize,
		locker:  new(sync.Mutex),
	}
}

// Size returns the current size of the pool.
func (p *Pool) Size() int {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.size
}

// Get gets an resource from the pool.
//
// It doesn't block if the pool is empty.
// Instead, it calls the Generator to generate a new resource to return.
//
// The Generator should not block. It blocks all pool operations.
//
// The Generator can be nil iff the pool is not empty.
func (p *Pool) Get(g Generator) interface{} {
	p.locker.Lock()
	defer p.locker.Unlock()
	if p.head == nil {
		return g()
	}
	ret := p.head
	p.head = ret.next
	p.size--
	if p.size == 0 {
		p.tail = nil
	}
	return ret.resource
}

// Put puts an resource into the pool.
//
// The return value indicates whether the resource has been put into the pool.
// It returns false iff the pool is already full.
func (p *Pool) Put(resource interface{}) bool {
	p.locker.Lock()
	defer p.locker.Unlock()
	if p.maxSize > 0 && p.size >= p.maxSize {
		return false
	}
	newItem := &node{
		resource: resource,
		next:     nil,
	}
	p.size++
	if p.size == 1 {
		p.head = newItem
		p.tail = newItem
		return true
	}
	p.tail.next = newItem
	p.tail = newItem
	return true
}
