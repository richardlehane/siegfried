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
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/persist"
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

func (t *testTree) String() string {
	str := "{TEST TREE Completes:"
	for i, v := range t.complete {
		str += v.String()
		if i < len(t.complete)-1 {
			str += ", "
		}
	}
	if len(t.incomplete) < 1 {
		return str + "}"
	}
	str += " Incompletes:"
	for i, v := range t.incomplete {
		str += v.kf.String()
		if i < len(t.incomplete)-1 {
			str += ", "
		}
	}
	return str + "}"
}

// KeyFrames returns a list of all KeyFrameIDs that are included in the test tree, including completes and incompletes
func (t *testTree) keyFrames() []keyFrameID {
	ret := make([]keyFrameID, len(t.complete), len(t.complete)+len(t.incomplete))
	copy(ret, t.complete)
	for _, v := range t.incomplete {
		ret = append(ret, v.kf)
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
	return
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

// TODO: This recursive function can overload the stack. Replace with a linear goroutine approach
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
		var success bool
		var offs []int
		if rev {
			success, offs = t.MatchR(b[:len(b)-o])
		} else {
			success, offs = t.Match(b[o:])
		}
		if success {
			for i, _ := range offs {
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
