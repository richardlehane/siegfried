// +build !go1.3

package wac

import "sync"

// pool of precons - just a simple free list
type pool struct {
	mu       *sync.Mutex
	template []int
	head     *item
}

type item struct {
	next *item
	p    precons
}

func newPool(s []Seq) *pool {
	t := make([]int, len(s))
	for i := range s {
		t[i] = len(s[i].Choices)
	}
	return &pool{&sync.Mutex{}, t, nil}
}

func (p *pool) get() precons {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.head == nil {
		return newPrecons(p.template)
	}
	precons := clear(p.head.p)
	p.head = p.head.next
	return precons
}

func (p *pool) put(v precons) {
	p.mu.Lock()
	p.head = &item{p.head, v}
	p.mu.Unlock()
}
