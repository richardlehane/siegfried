// Copyright 2018 Richard Lehane. All rights reserved.
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

// This file implements Aho Corasick searching for the bytematcher package
package bytematcher

import (
	//"fmt"
	"io"
	//"sort"
	//"strings"
	//"sync"

	"github.com/richardlehane/siegfried/internal/siegreader"
)

type searcher struct {
	bofWac        ac
	eofWac        ac
	maxBof        int
	maxEof        int
	wildSeqs      []seq          // separate out wild sequences to create a dynamic searcher for wildcard matching
	entanglements []entanglement // same len as wildSeqs
}

func newSearcher(bofSeqs []seq, eofSeqs []seq, entanglements map[int]entanglement) *searcher {
	return nil
}

func (s *searcher) search(buf *siegreader.Buffer) chan result {
	output := make(chan result)
	// check bof
	// check eof
	// build wild matcher
	// check wild bof
	return output
}

/* Index returns a channel of results, these contain the indexes (a double index: index of the Seq and index of the Choice)
// and offsets (in the input byte slice) of matching sequences.
func (ac *achm) index(input io.ByteReader) chan result {
	output := make(chan result)
	go ac.match(input, output)
	return output
}
*/
func (ac *achm) match(input io.ByteReader, results chan seqmatch) {
	var offset int64
	var progressResult = seqmatch{index: [2]int{-1, -1}}
	precons := ac.p.get()
	curr := ac.zero
	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		if trans := curr.transit[c]; trans != nil {
			curr = trans
		} else {
			for curr != ac.root {
				curr = curr.fail
				if trans := curr.transit[c]; trans != nil {
					curr = trans
					break
				}

			}
		}
		if curr.output != nil && (curr.outMax == -1 || curr.outMax >= offset-int64(curr.outMaxL)) {
			for _, o := range curr.output {
				if o.max == -1 || o.max >= offset-int64(o.length) {
					if o.subIndex == 0 || (precons[o.seqIndex][o.subIndex-1] != 0 && offset-int64(o.length) >= precons[o.seqIndex][o.subIndex-1]) {
						if precons[o.seqIndex][o.subIndex] == 0 {
							precons[o.seqIndex][o.subIndex] = offset
						}
						results <- seqmatch{index: [2]int{o.seqIndex, o.subIndex}, offset: offset - int64(o.length), length: o.length}
					}
				}
			}
		}
		if offset&(^offset+1) == offset && offset >= 1024 { // send powers of 2 greater than 512
			progressResult.offset = offset
			results <- progressResult
		}
	}
	ac.p.put(precons)
	close(results)
}

/* Index returns a channel of results, these contain the indexes (a double index: index of the Seq and index of the Choice)
// and offsets (in the input byte slice) of matching sequences.
func (ac *aclm) index(input io.ByteReader) chan result {
	output := make(chan result)
	go ac.match(input, output)
	return output
}
*/
func (ac *aclm) match(input io.ByteReader, results chan seqmatch) {
	var offset int64
	var progressResult = seqmatch{index: [2]int{-1, -1}}
	precons := ac.p.get()
	curr := ac.root
	for c, err := input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		if trans := curr.transit.get(c); trans != nil {
			curr = trans
		} else {
			for curr != ac.root {
				curr = curr.fail
				if trans := curr.transit.get(c); trans != nil {
					curr = trans
					break
				}

			}
		}
		if curr.output != nil && (curr.outMax == -1 || curr.outMax >= offset-int64(curr.outMaxL)) {
			for _, o := range curr.output {
				if o.max == -1 || o.max >= offset-int64(o.length) {
					if o.subIndex == 0 || (precons[o.seqIndex][o.subIndex-1] != 0 && offset-int64(o.length) >= precons[o.seqIndex][o.subIndex-1]) {
						if precons[o.seqIndex][o.subIndex] == 0 {
							precons[o.seqIndex][o.subIndex] = offset
						}
						results <- seqmatch{index: [2]int{o.seqIndex, o.subIndex}, offset: offset - int64(o.length), length: o.length}
					}
				}
			}
		}
		if offset&(^offset+1) == offset && offset >= 1024 { // send powers of 2 greater than 512
			progressResult.offset = offset
			results <- progressResult
		}
	}
	ac.p.put(precons)
	close(results)
}
