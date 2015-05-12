// Copyright 2015 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package siegreader

import "sync"

// pool of precons - just a simple free list
type pool struct {
	mu   *sync.Mutex
	fn   func() interface{}
	head *item
}

type item struct {
	next *item
	val  interface{}
}

func newPool(f func() interface{}) *pool {
	return &pool{
		mu: &sync.Mutex{},
		fn: f,
	}
}

func (p *pool) get() interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.head == nil {
		return p.fn()
	}
	ret := p.head.val
	p.head = p.head.next
	return ret
}

func (p *pool) put(v interface{}) {
	p.mu.Lock()
	p.head = &item{p.head, v}
	p.mu.Unlock()
}
