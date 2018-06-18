// +build go1.3,!appengine

package wac

import "sync"

// pool of precons - just an embedded sync.Pool
type pool struct{ *sync.Pool }

func preconsFn(s []Seq) func() interface{} {
	t := make([]int, len(s))
	for i := range s {
		t[i] = len(s[i].Choices)
	}
	return func() interface{} {
		return newPrecons(t)
	}
}

func newPool(s []Seq) *pool {
	return &pool{&sync.Pool{New: preconsFn(s)}}
}

func (p *pool) get() precons {
	return p.Get().(precons)
}

func (p *pool) put(v precons) {
	p.Put(clear(v)) // zero it when we put it back
}
