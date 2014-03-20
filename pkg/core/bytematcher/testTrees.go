package bytematcher

// Test trees link byte sequence and frame matches (from the sequence and frame sets) to keyframes. This link is sometimes direct if there are no
// further test to perform. Follow-up tests may be required to the left or to the right of the match.

type testTree struct {
	Complete         []keyframeID
	Incomplete       []followUp
	MaxLeftDistance  int
	MaxRightDistance int
	Left             []*testNode
	Right            []*testNode
}

type followUp struct {
	Kf keyframeID
	L  bool // have a left test
	R  bool // have a right test
}

type followupMatch struct {
	followUp  int
	distances []int
}

type testNode struct {
	Frame
	Success []int // followUp id
	Tests   []*testNode
}

func newTestNode(f Frame) *testNode {
	t := new(testNode)
	t.Frame = f
	t.Success = make([]int, 0)
	t.Tests = make([]*testNode, 0)
	return t
}

func hasTest(t []*testNode, f Frame) (*testNode, bool) {
	for _, nt := range t {
		if nt.Frame.Equals(f) {
			return nt, true
		}
	}
	return nil, false
}

func appendTests(ts []*testNode, f []Frame, fu int) []*testNode {
	if len(f) < 1 {
		return ts
	}
	nts := make([]*testNode, len(ts))
	copy(nts, ts)
	var t *testNode
	if nt, ok := hasTest(nts, f[0]); ok {
		t = nt
	} else {
		t = newTestNode(f[0])
		nts = append(nts, t)
	}
	if len(f) > 1 {
		for _, f1 := range f[1:] {
			if nt, ok := hasTest(t.Tests, f1); ok {
				t = nt
			} else {
				nt := newTestNode(f1)
				t.Tests = append(t.Tests, nt)
				t = nt
			}
		}
	}
	t.Success = append(t.Success, fu)
	return nts
}

func newTestTree() *testTree {
	return &testTree{
		make([]keyframeID, 0),
		make([]followUp, 0),
		0, 0,
		make([]*testNode, 0),
		make([]*testNode, 0),
	}
}

func (t *testTree) add(kf keyframeID, l []Frame, r []Frame) {
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
	t.Incomplete = append(t.Incomplete, followUp{kf, fl, fr})
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
	_, l := t.Frame.Length()
	return l + t.Frame.Max()
}

func maxLength(ts []*testNode) int {
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

func matchTestNodes(ts []*testNode, b []byte, rev bool) []followupMatch {
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
