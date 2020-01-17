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

func singleLen(f Frame) bool {
	min, max := f.Length()
	if min == max {
		return true
	}
	return false
}

// Blockify takes a signature segment, identifies any blocks within (frames linked by fixed offsets),
// converts those frames to block patterns within window frames (the window frame of the first frame in the block),
// but with a new length), and returns a new segment.
// If no blocks are within a segment, the original segment will be returned.
func Blockify(seg Signature) Signature {
	if len(seg) < 2 {
		return seg
	}
	ret := make(Signature, 0, len(seg))
	lst := seg[0]
	blk := []Frame{lst}
	for _, f := range seg[1:] {
		if lnk, _, _ := f.Linked(lst, -1, 0); lnk && singleLen(lst) && singleLen(f) {
			blk = append(blk, f)
		} else {
			ret = append(ret, blockify(blk))
			blk = []Frame{f}
		}
		lst = f
	}
	return append(ret, blockify(blk))
}

func blockify(seg []Frame) Frame {
	if len(seg) == 1 {
		return seg[0]
	}
	// identify Key by looking for longest Sequence Pattern within the segment
	var kf, kfl int
	for i, f := range seg {
		if _, ok := f.Pattern.(patterns.Sequence); ok { // we want to BMH the key, so this will only work on seqs
			l, _ := f.Length()
			if l > kfl {
				kfl = l
				kf = i
			}
		}
	}
	blk := &Block{}
	var fr Frame
	// Frame is the first frame in a BOF/PREV segment, or the last if a EOF/SUCC segment
	typ := Signature(seg).Characterise()
	// BMHify the Key and populate (switching) the L and R frames
	if typ <= Prev {
		fr = seg[0]
		blk.Key = patterns.BMH(seg[kf].Pattern, false)
		if kf < len(seg)-1 {
			blk.R = seg[kf+1:]
		}
		blk.L = make([]Frame, kf)
		for i := 0; i < kf; i++ {
			blk.L[i] = SwitchFrame(seg[i+1], seg[i].Pattern)
		}
	} else {
		fr = seg[len(seg)-1]
		blk.Key = patterns.BMH(seg[kf].Pattern, true)
		if kf > 0 {
			blk.L = seg[:kf]
		}
		blk.R = make([]Frame, len(seg)-kf-1)
		idx := len(blk.R) - 1
		for i := len(seg) - 1; i > kf; i-- {
			blk.R[idx] = SwitchFrame(seg[i-1], seg[i].Pattern)
			idx--
		}
	}
	// calc block length by tallying TotalLength of L and R frames plus length of the pattern
	blk.Le, _ = blk.Key.Length()
	for _, f := range blk.L {
		blk.Le += TotalLength(f)
		blk.Off += TotalLength(f)
	}
	for _, f := range blk.R {
		blk.Le += TotalLength(f)
		blk.OffR += TotalLength(f)
	}
	fr.Pattern = blk
	return fr
}

// Block combines Frames that are linked to each other by a fixed offset into a single Pattern
// Patterns within a block must have a single length (i.e. no Choice patterns with varying lengths).
// Blocks are used within the Machine pattern to cluster frames to identify repetitions & optimise searching.
type Block struct {
	L    []Frame
	R    []Frame
	Key  patterns.Pattern
	Le   int // Pattern length
	Off  int // fixed offset of the Key, relative to the first frame in the block
	OffR int // fixed offset of the Key, relative to the last frame in the block
}

func (bl *Block) Test(b []byte) ([]int, int) {
	if bl.Off >= len(b) {
		return nil, 0
	}
	ls, jmp := bl.Key.Test(b[bl.Off:])
	if len(ls) < 1 {
		return nil, jmp
	}
	ld := bl.Off
	for i := len(bl.L) - 1; i >= 0; i-- {
		if ld < 0 {
			return nil, jmp
		}
		j, _ := bl.L[i].MatchNR(b[:ld], 0)
		if j < 0 {
			return nil, jmp
		}
		ld -= j
	}
	rd := bl.Off + ls[0]
	for _, rf := range bl.R {
		if rd > len(b)-1 {
			return nil, jmp
		}
		j, _ := rf.MatchN(b[rd:], 0)
		if j < 0 {
			return nil, jmp
		}
		rd += j
	}
	return []int{bl.Le}, jmp
}

func (bl *Block) TestR(b []byte) ([]int, int) {
	if bl.OffR >= len(b) {
		return nil, 0
	}
	ls, jmp := bl.Key.TestR(b[:len(b)-bl.OffR])
	if len(ls) < 1 {
		return nil, jmp
	}
	ld := bl.OffR + ls[0]
	for i := len(bl.L) - 1; i >= 0; i-- {
		if ld < 0 {
			return nil, jmp
		}
		j, _ := bl.L[i].MatchNR(b[:ld], 0)
		if j < 0 {
			return nil, jmp
		}
		ld -= j
	}
	rd := len(b) - bl.OffR
	for _, rf := range bl.R {
		if rd > len(b)-1 {
			return nil, jmp
		}
		j, _ := rf.MatchN(b[rd:], 0)
		if j < 0 {
			return nil, jmp
		}
		rd += j
	}
	return []int{bl.Le}, jmp
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
		bl.Le != bl2.Le || bl.Off != bl2.Off || bl.OffR != bl2.OffR {
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
	return bl.Le, bl.Le
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
	ls.SaveSmallInt(len(bl.R))
	for _, f := range bl.R {
		f.Save(ls)
	}
	bl.Key.Save(ls)
	ls.SaveInt(bl.Le)
	ls.SaveInt(bl.Off)
	ls.SaveInt(bl.OffR)
}

func loadBlock(ls *persist.LoadSaver) patterns.Pattern {
	bl := &Block{}
	bl.L = make([]Frame, ls.LoadSmallInt())
	for i := range bl.L {
		bl.L[i] = Load(ls)
	}
	bl.R = make([]Frame, ls.LoadSmallInt())
	for i := range bl.R {
		bl.R[i] = Load(ls)
	}
	bl.Key = patterns.Load(ls)
	bl.Le = ls.LoadInt()
	bl.Off = ls.LoadInt()
	bl.OffR = ls.LoadInt()
	return bl
}
