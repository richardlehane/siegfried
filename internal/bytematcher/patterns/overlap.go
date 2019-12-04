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

// func Overlap calculates the max distance before a possible overlap with multiple matches the same Pattern
// e.g. 0xAABBAA has a length of 3 but returns 2
func Overlap(p Pattern) int {
	seqs := p.Sequences()
	if len(seqs) < 1 {
		return 1
	}
	ret := len(seqs[0])
	for _, v := range seqs {
		for _, vv := range seqs {

		}
	}
	return ret
}

func overlap(a []byte, b []byte) int {
	ret := 1
	for ; ret < len(a) && ret < len(b); ret++ {
		for i := ret; i < len(a) && i < len(b); i++ {
			if a[i] != b[i-1] {

			}
		}
	}
	return ret
}
