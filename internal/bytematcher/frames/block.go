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

package frames

import (
	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	"github.com/richardlehane/siegfried/internal/persist"
)

// Blockify takes a signature segment, identifies any blocks within (frames linked by fixed offsets),
// converts those frames to block patterns within window frames (the window frame of the first frame in the block),
// but with a new length), and returns a new segment.
// If no blocks are within a segment, the original segment will be returned.
func Blockify(seg Signature) Signature {
	if len(seg) < 2 {
		return seg
	}
	ret := make(Signature, 0, len(seg))
	var blk []Frame
	lst := seg[0]
	for _, f := range seg[1:] {
		if lnk, _, _ := f.Linked(lst, -1, 0); lnk {
			if len(blk) == 0 {
				blk = append(blk, lst, f)
			} else {
				blk = append(blk, f)
			}
		} else {
			if len(blk) > 0 {
				if len(blk) > 0 {
					ret = append(ret, blockify(blk))
				}
				blk = []Frame{}
			}
			ret = append(ret, f)
		}
	}
	if len(blk) > 0 {
		ret = append(ret, blockify(blk))
	}
	return ret
}

func blockify(seg []Frame) Frame {
	return seg[0]
}

// Block combines Frames that are linked to each other by a fixed offset into a single Pattern
// Blocks are used within the Machine pattern to cluster frames to identify repetitions & optimise searching.
type Block struct {
	L    []Frame
	R    []Frame
	Key  patterns.Pattern
	Min  int // Min pattern length
	Max  int // Max pattern length
	Off  int // fixed offset of the Key, relative to the first frame in the block
	OffR int // fixed offset of the Key, relative to the last frame in the block
}

func (bl *Block) Test(b []byte) ([]int, int) {
	if bl.Off >= len(b) {
		return nil, 0
	}
	l, jmp := bl.Key.Test(b[bl.Off:])
	if len(l) < 1 {
		return l, jmp
	}
	return l, jmp
}

func (bl *Block) TestR(b []byte) ([]int, int) {
	return nil, 1
}

func (bl *Block) Equals(pat patterns.Pattern) bool {
	bl2, ok := pat.(*Block)
	if !ok {
		return false
	}
	if !bl.Key.Equals(bl2.Key) {
		return false
	}
	if len(bl.L) != len(bl2.L) || len(bl.R) != len(bl2.R) ||
		bl.Min != bl2.Min || bl.Max != bl2.Max ||
		bl.Off != bl2.Off || bl.OffR != bl2.OffR {
		return false
	}
	for i, v := range bl.L {
		if !v.Equals(bl2.L[i]) {
			return false
		}
	}
	for i, v := range bl.R {
		if !v.Equals(bl2.R[i]) {
			return false
		}
	}
	return true
}

func (bl *Block) Length() (int, int) {
	return bl.Min, bl.Max
}

// Blocks are used where sequence matching inefficient
func (bl *Block) NumSequences() int              { return 0 }
func (bl *Block) Sequences() []patterns.Sequence { return nil }

func (bl *Block) String() string {
	str := bl.Key.String()
	if len(bl.L) > 0 {
		str += "; L:"
	}
	for i, v := range bl.L {
		if i > 0 {
			str += " | "
		}
		str += v.String()
	}
	if len(bl.R) > 0 {
		str += "; R:"
	}
	for i, v := range bl.R {
		if i > 0 {
			str += " | "
		}
		str += v.String()
	}
	return "b {" + str + "}"
}

func (bl *Block) Save(ls *persist.LoadSaver) {
	ls.SaveByte(blockLoader)
	ls.SaveSmallInt(len(bl.L))
	for _, f := range bl.L {
		f.Save(ls)
	}
}

func loadBlock(ls *persist.LoadSaver) patterns.Pattern {
	return &Block{}
}
