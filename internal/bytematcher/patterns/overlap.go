// Copyright 2019 Richard Lehane. All rights reserved.
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

package patterns

// func Overlap calculates the max distance before a possible overlap with multiple matches of the same Pattern
// e.g. 0xAABBAA has a length of 3 but returns 2
func Overlap(p Pattern) int {
	return aggregateOverlap(p, overlap)
}

// func OverlapR calculates the max distance before a possible overlap with multiple matches of the same Pattern,
// matching in reverse
// e.g. EOFE has a length of 4 but returns 3
func OverlapR(p Pattern) int {
	return aggregateOverlap(p, overlapR)
}

func aggregateOverlap(p Pattern, of overlapFunc) int {
	var counter int // prevent combinatorial explosions
	seqs := p.Sequences()
	if len(seqs) < 1 {
		return 1
	}
	ret := len(seqs[0])
	for _, v := range seqs {
		for _, vv := range seqs {
			counter++
			if counter > 1000000 {
				//panic("counter exploded!")
				return ret
			}
			if r := of(v, vv); r < ret {
				ret = r
			}
		}
	}
	return ret
}

type overlapFunc func([]byte, []byte) int

func overlap(a, b []byte) int {
	var ret int = 1
	for ; ret < len(a); ret++ {
		success := true
		for i := 0; ret+i < len(a) && i < len(b); i++ {
			if a[ret+i] != b[i] {
				success = false
				break
			}
		}
		if success {
			break
		}
	}
	return ret
}

func overlapR(a, b []byte) int {
	var ret int = 1
	for ; ret < len(a); ret++ {
		success := true
		for i := 0; ret+i < len(a) && i < len(b); i++ {
			if a[len(a)-ret-i-1] != b[len(b)-i-1] {
				success = false
				break
			}
		}
		if success {
			break
		}
	}
	return ret
}
