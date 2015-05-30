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

package bytematcher

import (
	"bytes"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/core/signature"
)

// Sequence Sets and Frame Sets

// As far as possible, signatures are flattened into simple byte sequences grouped into two sets: BOF and EOF sets.
// When a byte sequence is matched, the TestTree is examined for keyframe matches and to conduct further tests.
type seqSet struct {
	set           []wac.Seq
	testTreeIndex []int // The index of the testTree for the first choices. For subsequence choices, add the index of that choice to the test tree index.
}

func (ss *seqSet) save(ls *signature.LoadSaver) {
	ls.SaveSmallInt(len(ss.set))
	for _, v := range ss.set {
		ls.SaveBigInts(v.MaxOffsets)
		ls.SaveSmallInt(len(v.Choices))
		for _, w := range v.Choices {
			ls.SaveSmallInt(len(w))
			for _, x := range w {
				ls.SaveBytes(x)
			}
		}
	}
	ls.SaveInts(ss.testTreeIndex)
}

func loadSeqSet(ls *signature.LoadSaver) *seqSet {
	ret := &seqSet{}
	le := ls.LoadSmallInt()
	if le == 0 {
		_ = ls.LoadInts() // discard the empty testtreeindex list too
		return ret
	}
	ret.set = make([]wac.Seq, le)
	for i := range ret.set {
		ret.set[i].MaxOffsets = ls.LoadBigInts()
		ret.set[i].Choices = make([]wac.Choice, ls.LoadSmallInt())
		for j := range ret.set[i].Choices {
			ret.set[i].Choices[j] = make(wac.Choice, ls.LoadSmallInt())
			for k := range ret.set[i].Choices[j] {
				ret.set[i].Choices[j][k] = ls.LoadBytes()
			}
		}
	}
	ret.testTreeIndex = ls.LoadInts()
	return ret
}

// helper funcs to test equality of wac.Seq
func choiceExists(a []byte, b wac.Choice) bool {
	for _, v := range b {
		if bytes.Equal(a, v) {
			return true
		}
	}
	return false
}

func seqEquals(a wac.Seq, b wac.Seq) bool {
	if len(a.MaxOffsets) != len(b.MaxOffsets) || len(a.Choices) != len(b.Choices) {
		return false
	}
	for i := range a.MaxOffsets {
		if a.MaxOffsets[i] != b.MaxOffsets[i] {
			return false
		}
	}
	for i := range a.Choices {
		if len(a.Choices[i]) != len(b.Choices[i]) {
			return false
		}
		for _, v := range a.Choices[i] {
			if !choiceExists(v, b.Choices[i]) {
				return false
			}
		}
	}
	return true
}

func (ss *seqSet) exists(seq wac.Seq) (int, bool) {
	for i, v := range ss.set {
		if seqEquals(seq, v) {
			return i, true
		}
	}
	return -1, false
}

// Add sequence to set. Provides latest testTreeIndex, returns actual testTreeIndex for hit insertion.
func (ss *seqSet) add(seq wac.Seq, hi int) int {
	i, ok := ss.exists(seq)
	if ok {
		return ss.testTreeIndex[i]
	}
	ss.set = append(ss.set, seq)
	ss.testTreeIndex = append(ss.testTreeIndex, hi)
	return hi
}

// Some signatures cannot be represented by simple byte sequences. The first or last frames from these sequences are added to the BOF or EOF frame sets.
// Like sequences, frame matches are referred to the TestTree for further testing.
type frameSet struct {
	set           []frames.Frame
	testTreeIndex []int
}

func (fs *frameSet) save(ls *signature.LoadSaver) {
	ls.SaveSmallInt(len(fs.set))
	for _, f := range fs.set {
		f.Save(ls)
	}
	ls.SaveInts(fs.testTreeIndex)
}

func loadFrameSet(ls *signature.LoadSaver) *frameSet {
	ret := &frameSet{}
	le := ls.LoadSmallInt()
	if le == 0 {
		_ = ls.LoadInts()
		return ret
	}
	ret.set = make([]frames.Frame, le)
	for i := range ret.set {
		ret.set[i] = frames.Load(ls)
	}
	ret.testTreeIndex = ls.LoadInts()
	return ret
}

// Add frame to set. Provides current testerIndex, returns actual testerIndex for hit insertion.
func (fs *frameSet) add(f frames.Frame, hi int) int {
	for i, f1 := range fs.set {
		if f1.Equals(f) {
			return fs.testTreeIndex[i]
		}
	}
	fs.set = append(fs.set, f)
	fs.testTreeIndex = append(fs.testTreeIndex, hi)
	return hi
}

type fsmatch struct {
	idx    int
	off    int64
	length int
}

func (fs *frameSet) index(buf siegreader.Buffer, rev bool, quit chan struct{}) chan fsmatch {
	ret := make(chan fsmatch)
	go func() {
		var i int
		for {
			select {
			case <-quit:
				close(ret)
				return
			default:
			}
			if i >= len(fs.set) {
				close(ret)
				return
			}
			f := fs.set[i]
			var match bool
			var matches []int
			if rev {
				slc, err := buf.EofSlice(0, frames.TotalLength(f))
				if err != nil {
					close(ret)
					return
				}
				match, matches = f.MatchR(slc)
			} else {
				slc, err := buf.Slice(0, frames.TotalLength(f))
				if err != nil {
					close(ret)
					return
				}
				match, matches = f.Match(slc)
			}
			if match {
				var min int
				if !rev {
					min, _ = f.Length()
				}
				for _, off := range matches {
					ret <- fsmatch{i, int64(off - min), min}
				}
			}
			i++
		}
		close(ret)
	}()
	return ret
}
