// Copyright 2014 Richard Lehane. All rights reserved.
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

// +build !go1.3

package siegreader

type pool struct {
	fn   func() interface{}
	vals chan interface{}
}

func newPool(f func() interface{}) *pool {
	return &pool{
		fn:   f,
		vals: make(chan interface{}, 5),
	}
}

func (p *pool) Get() interface{} {
	select {
	case v := <-p.vals:
		return v
	default:
		return p.fn()
	}
}

func (p *pool) Put(v interface{}) {
	select {
	case p.vals <- v:
	default:
	}
}

func New() *Buffers {
	return &Buffers{
		newPool(newStream),
		newPool(newFile),
		&datas{
			newPool(newBigFile),
			newPool(newSmallFile),
			newPool(newMmap),
		},
	}
}

type Buffers struct {
	spool  *pool // Pool of stream Buffers
	fpool  *pool // Pool of file Buffers
	fdatas *datas
}

// Data pool (used by file)
type datas struct {
	bfpool *pool
	sfpool *pool
	mpool  *pool
}
