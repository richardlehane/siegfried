package process

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

func TestMaxLength(t *testing.T) {
	test := newTestTree()
	test.add([2]int{0, 0}, []frames.Frame{}, []frames.Frame{frames.TestFrames[0], frames.TestFrames[3], frames.TestFrames[6]})
	test.add([2]int{0, 0}, []frames.Frame{}, []frames.Frame{frames.TestFrames[1], frames.TestFrames[3]})
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
