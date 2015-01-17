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

// +build go1.3

package siegreader

import "sync"

func New() *Buffers {
	return &Buffers{
		&sync.Pool{
			New: newStream,
		},
		&sync.Pool{
			New: newFile,
		},
		&datas{
			&sync.Pool{
				New: newBigFile,
			},
			&sync.Pool{
				New: newSmallFile,
			},
			&sync.Pool{
				New: newMmap,
			},
		},
	}
}

type Buffers struct {
	spool  *sync.Pool // Pool of stream Buffers
	fpool  *sync.Pool // Pool of file Buffers
	fdatas *datas
}

// Data pool (used by file)
type datas struct {
	bfpool *sync.Pool
	sfpool *sync.Pool
	mpool  *sync.Pool
}
