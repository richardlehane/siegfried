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

package process

import "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"

// Test trees link byte sequence and frame matches (from the sequence and frame sets) to keyframes. This link is sometimes direct if there are no
// further test to perform. Follow-up tests may be required to the left or to the right of the match.

type testTree struct {
	Complete         []KeyFrameID
	Incomplete       []FollowUp
	MaxLeftDistance  int
	MaxRightDistance int
	Left             []*testNode
	Right            []*testNode
}

func (t *testTree) String() string {
	str := "{TEST TREE Completes:"
	for i, v := range t.Complete {
		str += v.String()
		if i < len(t.Complete)-1 {
			str += ", "
		}
	}
	if len(t.Incomplete) < 1 {
		return str + "}"
	}
	str += " Incompletes:"
	for i, v := range t.Incomplete {
		str += v.Kf.String()
		if i < len(t.Incomplete)-1 {
			str += ", "
		}
	}
	return str + "}"
}

type FollowUp struct {
	Kf KeyFrameID
	L  bool // have a left test
	R  bool // have a right test
}

type followupMatch struct {
	FollowUp  int
	Distances []int
}

type testNode struct {
	frames.Frame
	Success []int // followUp id
	Tests   []*testNode
}

func newtestNode(f frames.Frame) *testNode {
	t := new(testNode)
	t.Frame = f
	t.Success = make([]int, 0)
	t.Tests = make([]*testNode, 0)
	return t
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
			if nt, ok := hasTest(t.Tests, f1); ok {
				t = nt
			} else {
				nt := newtestNode(f1)
				t.Tests = append(t.Tests, nt)
				t = nt
			}
		}
	}
	t.Success = append(t.Success, fu)
	return nts
}

func (t *testTree) add(kf KeyFrameID, l []frames.Frame, r []frames.Frame) {
	if len(l) == 0 && len(r) == 0 {
		t.Complete = append(t.Complete, kf)
		return
	}
	var fl, fr bool
	if len(l) > 0 {
		fl = true
	}
	if len(r) > 0 {
		fr = true
	}
	t.Incomplete = append(t.Incomplete, FollowUp{kf, fl, fr})
	fu := len(t.Incomplete) - 1
	if fl {
		t.Left = appendTests(t.Left, l, fu)
	}
	if fr {
		t.Right = appendTests(t.Right, r, fu)
	}
	return
}

func (t *testNode) length() int {
	return frames.TotalLength(t.Frame)
}

func MaxLength(ts []*testNode) int {
	var max int
	var delve func(t *testNode, this int)
	delve = func(t *testNode, this int) {
		if len(t.Tests) == 0 {
			if this+t.length() > max {
				max = this + t.length()
			}
		}
		for _, nt := range t.Tests {
			delve(nt, this+t.length())
		}
	}
	for _, t := range ts {
		delve(t, 0)
	}
	return max
}

func MatchTestNodes(ts []*testNode, b []byte, rev bool) []followupMatch {
	ret := make([]followupMatch, 0)
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
			for _, s := range t.Success {
				ret = append(ret, followupMatch{s, offs})
			}
			for _, off := range offs {
				for _, test := range t.Tests {
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
