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
	"sort"

	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/persist"
)

// Test trees link byte sequence and frame matches (from the sequence and frame sets) to keyframes. This link is sometimes direct if there are no
// further test to perform. Follow-up tests may be required to the left or to the right of the match.
type testTree struct {
	complete         []keyFrameID
	incomplete       []followUp
	maxLeftDistance  int
	maxRightDistance int
	left             []*testNode
	right            []*testNode
}

func saveTests(ls *persist.LoadSaver, tts []*testTree) {
	ls.SaveSmallInt(len(tts))
	for _, tt := range tts {
		ls.SaveSmallInt(len(tt.complete))
		for _, kfid := range tt.complete {
			ls.SaveSmallInt(kfid[0])
			ls.SaveSmallInt(kfid[1])
		}
		ls.SaveSmallInt(len(tt.incomplete))
		for _, fu := range tt.incomplete {
			ls.SaveSmallInt(fu.kf[0])
			ls.SaveSmallInt(fu.kf[1])
			ls.SaveBool(fu.l)
			ls.SaveBool(fu.r)
		}
		ls.SaveInt(tt.maxLeftDistance)
		ls.SaveInt(tt.maxRightDistance)
		saveTestNodes(ls, tt.left)
		saveTestNodes(ls, tt.right)
	}
}

func loadTests(ls *persist.LoadSaver) []*testTree {
	l := ls.LoadSmallInt()
	ret := make([]*testTree, l)
	for i := range ret {
		ret[i] = &testTree{}
		ret[i].complete = make([]keyFrameID, ls.LoadSmallInt())
		for j := range ret[i].complete {
			ret[i].complete[j][0] = ls.LoadSmallInt()
			ret[i].complete[j][1] = ls.LoadSmallInt()
		}
		ret[i].incomplete = make([]followUp, ls.LoadSmallInt())
		for j := range ret[i].incomplete {
			ret[i].incomplete[j].kf[0] = ls.LoadSmallInt()
			ret[i].incomplete[j].kf[1] = ls.LoadSmallInt()
			ret[i].incomplete[j].l = ls.LoadBool()
			ret[i].incomplete[j].r = ls.LoadBool()
		}
		ret[i].maxLeftDistance = ls.LoadInt()
		ret[i].maxRightDistance = ls.LoadInt()
		ret[i].left = loadTestNodes(ls)
		ret[i].right = loadTestNodes(ls)
	}
	return ret
}

type followUp struct {
	kf keyFrameID
	l  bool // have a left test
	r  bool // have a right test
}

type followupMatch struct {
	followUp  int
	distances []int
}

type testNode struct {
	frames.Frame
	success []int // followUp id
	tests   []*testNode
}

func saveTestNodes(ls *persist.LoadSaver, tns []*testNode) {
	ls.SaveSmallInt(len(tns))
	for _, n := range tns {
		n.Frame.Save(ls)
		ls.SaveInts(n.success)
		saveTestNodes(ls, n.tests)
	}
}

func loadTestNodes(ls *persist.LoadSaver) []*testNode {
	l := ls.LoadSmallInt()
	if l == 0 {
		return nil
	}
	ret := make([]*testNode, l)
	for i := range ret {
		ret[i] = &testNode{
			frames.Load(ls),
			ls.LoadInts(),
			loadTestNodes(ls),
		}
	}
	return ret
}

func newtestNode(f frames.Frame) *testNode {
	return &testNode{
		Frame: f,
	}
}

func hasTest(t []*testNode, f frames.Frame) (*testNode, bool) {
	for _, nt := range t {
		if nt.Frame.Equals(f) {
			return nt, true
		}
	}
	return nil, false
}

func appendTests(ts []*testNode, f []frames.Frame, fu int) []*testNode {
	// if our signature segment is empty just return ts
	if len(f) < 1 {
		return ts
	}
	nts := make([]*testNode, len(ts))
	copy(nts, ts)
	var t *testNode
	if nt, ok := hasTest(nts, f[0]); ok {
		t = nt
	} else {
		t = newtestNode(f[0])
		nts = append(nts, t)
	}
	if len(f) > 1 {
		for _, f1 := range f[1:] {
			if nt, ok := hasTest(t.tests, f1); ok {
				t = nt
			} else {
				nt := newtestNode(f1)
				t.tests = append(t.tests, nt)
				t = nt
			}
		}
	}
	t.success = append(t.success, fu)
	return nts
}

func (t *testTree) add(kf keyFrameID, l []frames.Frame, r []frames.Frame) {
	if len(l) == 0 && len(r) == 0 {
		t.complete = append(t.complete, kf)
		return
	}
	var fl, fr bool
	if len(l) > 0 {
		fl = true
	}
	if len(r) > 0 {
		fr = true
	}
	t.incomplete = append(t.incomplete, followUp{kf, fl, fr})
	fu := len(t.incomplete) - 1
	if fl {
		t.left = appendTests(t.left, l, fu)
	}
	if fr {
		t.right = appendTests(t.right, r, fu)
	}
}

func (t *testNode) length() int {
	return frames.TotalLength(t.Frame)
}

func maxLength(ts []*testNode) int {
	var max int
	var delve func(t *testNode, this int)
	delve = func(t *testNode, this int) {
		if len(t.tests) == 0 {
			if this+t.length() > max {
				max = this + t.length()
			}
		}
		for _, nt := range t.tests {
			delve(nt, this+t.length())
		}
	}
	for _, t := range ts {
		delve(t, 0)
	}
	return max
}

/*
Consider adding new calculated values for maxLeftIter and maxRightIter. These would use the new MaxMatches methods on the Frames

	to determine the theoretical max times we'd have to iterate in order to generate all the possible followUp hits.
*/
func maxMatches(ts []*testNode, l int) int {
	if len(ts) == 0 || l == 0 {
		return 0
	}
	var iters int
	maxes := make(map[int]int)
	var delve func(t *testNode, this int)
	delve = func(t *testNode, this int) {
		if iters > 1000 {
			return
		}
		iters++
		mm, rem, min := t.MaxMatches(this)
		for mm > 0 {
			for _, fu := range t.success {
				maxes[fu]++
			}
			for _, nt := range t.tests {
				delve(nt, rem)
			}
			mm--
			rem = rem - min
		}
	}
	for _, t := range ts {
		delve(t, l)
	}
	if iters > 1000 {
		return iters
	}
	maxSlc := make([]int, len(maxes))
	var iter int
	for _, v := range maxes {
		maxSlc[iter] = v
		iter++
	}
	sort.Ints(maxSlc)
	return maxSlc[len(maxSlc)-1]
}

// TODO: This recursive function can overload the stack. Replace with a lazy approach
// Could it return a closure that itself returns one followupMatch per keyframe ID?
func matchTestNodes(ts []*testNode, b []byte, rev bool) []followupMatch {
	ret := []followupMatch{}
	if b == nil {
		return ret
	}
	var match func(t *testNode, o int)
	match = func(t *testNode, o int) {
		if o >= len(b) {
			return
		}
		var offs []int
		if rev {
			offs = t.MatchR(b[:len(b)-o])
		} else {
			offs = t.Match(b[o:])
		}
		if len(offs) > 0 {
			for i := range offs {
				offs[i] = offs[i] + o
			}
			for _, s := range t.success {
				ret = append(ret, followupMatch{s, offs})
			}
			for _, off := range offs {
				for _, test := range t.tests {
					match(test, off)
				}
			}
		}
	}
	for _, t := range ts {
		match(t, 0)
	}
	return ret
}

// KeyFrames returns a list of all KeyFrameIDs that are included in the test tree, including completes and incompletes
// Used in scorer.go
func (t *testTree) keyFrames() []keyFrameID {
	ret := make([]keyFrameID, len(t.complete), len(t.complete)+len(t.incomplete))
	copy(ret, t.complete)
	for _, v := range t.incomplete {
		ret = append(ret, v.kf)
	}
	return ret
}

// FilterTests returns indexes into the main slice of testTree, given a slice of keyframe IDs.
// Used in scorer.go to select a subset of sequences and tests for dynamic matching.
func filterTests(ts []*testTree, kfids []keyFrameID) []int {
	ret := make([]int, 0, len(kfids)) // will return length always equal kfids (no multiple kfids could attach to a single tt, so may be less)? would it be faster to do the outer loop on kfids. Current each tt can onlya appear once
outer:
	for idx, tt := range ts {
		for _, c := range tt.complete {
			for _, k := range kfids {
				if c == k {
					ret = append(ret, idx)
					continue outer
				}
			}
		}
		for _, ic := range tt.incomplete {
			for _, k := range kfids {
				if ic.kf == k {
					ret = append(ret, idx)
					continue outer
				}
			}
		}
	}
	return ret
}
