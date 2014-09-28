package process

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
)

// Shared test testNodes (exported so they can be used by the other bytematcher packages)
var TesttestNodes = []*testNode{
	&testNode{
		Frame:   tests.TestFrames[3],
		Success: []int{},
		Tests: []*testNode{
			&testNode{
				Frame:   tests.TestFrames[1],
				Success: []int{0},
				Tests:   []*testNode{},
			},
		},
	},
	&testNode{
		Frame:   tests.TestFrames[6],
		Success: []int{},
		Tests: []*testNode{
			&testNode{
				Frame:   tests.TestFrames[2],
				Success: []int{0},
				Tests:   []*testNode{},
			},
		},
	},
}

// Shared test testTree (exported so they can be used by the other bytematcher packages)
var TestTestTree = &testTree{
	Complete: []KeyFrameID{},
	Incomplete: []FollowUp{
		FollowUp{
			Kf: KeyFrameID{1, 0},
			L:  true,
			R:  true,
		},
	},
	MaxLeftDistance:  10,
	MaxRightDistance: 30,
	Left:             []*testNode{TesttestNodes[0]},
	Right:            []*testNode{TesttestNodes[1]},
}

func TestMaxLength(t *testing.T) {
	test := &testTree{}
	test.add([2]int{0, 0}, []frames.Frame{}, []frames.Frame{tests.TestFrames[0], tests.TestFrames[3], tests.TestFrames[6]})
	test.add([2]int{0, 0}, []frames.Frame{}, []frames.Frame{tests.TestFrames[1], tests.TestFrames[3]})
	if MaxLength(test.Right) != 33 {
		t.Errorf("maxLength fail: expecting 33 got %v", MaxLength(test.Right))
	}
}

func TestMatchLeft(t *testing.T) {
	left := MatchTestNodes(TestTestTree.Left, Sample[:8], true)
	if len(left) != 1 {
		t.Errorf("expecting one match, got %v", len(left))
	}
	if left[0].FollowUp != 0 {
		t.Errorf("expecting 0, got %v", left[0].FollowUp)
	}
}

func TestMatchRight(t *testing.T) {
	right := MatchTestNodes(TestTestTree.Right, Sample[8+5:], false)
	if len(right) != 1 {
		t.Errorf("expecting one match, got %v", len(right))
	}
	if right[0].FollowUp != 0 {
		t.Errorf("expecting 0, got %v", right[0].FollowUp)
	}
}
