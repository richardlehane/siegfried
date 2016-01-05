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

// bigfile handles files that are too large to mmap (normally encountered on 32-bit machines)
type bigfile struct {
	*file
	eof   [eofSz]byte
	wheel [wheelSz]byte

	mu               sync.Mutex
	i                int // wheel offset for next write
	start, end, last int64
}

func newBigFile() interface{} {
	return &bigfile{last: int64(readSz)}
}

func (bf *bigfile) setSource(f *file) {
	bf.file = f
	// fill the EOF slice
	bf.src.ReadAt(bf.eof[:], bf.sz-int64(eofSz))
}

func (bf *bigfile) reset() {
	bf.i = 0
	bf.last = 0
}

type enc uint8

const (
	notEnc enc = iota
	wheelEnc
	eofEnc
)

// is the requested slice in the wheel or the eof?
func (bf *bigfile) enclosed(o int64, l int) enc {
	if bf.sz-o <= int64(eofSz) {
		return eofEnc
	}
	if o >= bf.start && o+int64(l) <= bf.end {
		return wheelEnc
	}
	return notEnc
}

func (bf *bigfile) adjacent(o int64, l int) bool {
	if l != readSz {
		return false
	}
	if bf.last == o {
		return true
	}
	return false
}

func (bf *bigfile) progressSlice(o int64) []byte {
	if bf.i == 0 {
		bf.start = o
		i, _ := bf.src.Read(bf.wheel[:])
		bf.end = bf.start + int64(i)
		if i < readSz {
			return nil
		}
	}
	slc := bf.wheel[bf.i : bf.i+readSz]
	bf.i += readSz
	return slc
}

func (bf *bigfile) slice(o int64, l int) []byte {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	if bf.adjacent(o, l) {
		bf.last += int64(readSz)
		return bf.progressSlice(o)
	}
	ret := make([]byte, l)
	// if within the wheel copy
	switch bf.enclosed(o, l) {
	case eofEnc:
		x := eofSz - int(bf.sz-o)
		copy(ret, bf.eof[x:x+l])
		return ret
	case wheelEnc:
		copy(ret, bf.wheel[int(o-bf.start):int(o-bf.start)+l])
		return ret
	}
	// otherwise we just expose the underlying reader at
	bf.src.ReadAt(ret, o)
	return ret
}

func (bf *bigfile) eofSlice(o int64, l int) []byte {
	if o+int64(l) > int64(eofSz) {
		ret := make([]byte, l)
		bf.mu.Lock()
		defer bf.mu.Unlock()
		bf.src.ReadAt(ret, bf.sz-o-int64(l))
		return ret
	}
	return bf.eof[eofSz-int(o)-l : eofSz-int(o)]
}
