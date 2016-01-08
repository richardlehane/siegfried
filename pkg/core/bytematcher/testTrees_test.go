package bytematcher

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/persist"
)

var TesttestNodes = []*testNode{
	&testNode{
		Frame:   tests.TestFrames[3],
		success: []int{},
		tests: []*testNode{
			&testNode{
				Frame:   tests.TestFrames[1],
				success: []int{0},
				tests:   []*testNode{},
			},
		},
	},
	&testNode{
		Frame:   tests.TestFrames[6],
		success: []int{},
		tests: []*testNode{
			&testNode{
				Frame:   tests.TestFrames[2],
				success: []int{0},
				tests:   []*testNode{},
			},
		},
	},
}

var TestTestTree = &testTree{
	complete: []keyFrameID{},
	incomplete: []followUp{
		followUp{
			kf: keyFrameID{1, 0},
			l:  true,
			r:  true,
		},
	},
	maxLeftDistance:  10,
	maxRightDistance: 30,
	left:             []*testNode{TesttestNodes[0]},
	right:            []*testNode{TesttestNodes[1]},
}

func TestMaxLength(t *testing.T) {
	test := &testTree{}
	test.add([2]int{0, 0}, []frames.Frame{}, []frames.Frame{tests.TestFrames[0], tests.TestFrames[3], tests.TestFrames[6]})
	test.add([2]int{0, 0}, []frames.Frame{}, []frames.Frame{tests.TestFrames[1], tests.TestFrames[3]})
	saver := persist.NewLoadSaver(nil)
	saveTests(saver, []*testTree{test, test})
	loader := persist.NewLoadSaver(saver.Bytes())
	tests := loadTests(loader)
	test = tests[1]
	if maxLength(test.right) != 33 {
		t.Errorf("maxLength fail: expecting 33 got %v", maxLength(test.right))
	}
}

func TestMatchLeft(t *testing.T) {
	left := matchTestNodes(TestTestTree.left, Sample[:8], true)
	if len(left) != 1 {
		t.Errorf("expecting one match, got %v", len(left))
	}
	if left[0].followUp != 0 {
		t.Errorf("expecting 0, got %v", left[0].followUp)
	}
}

func TestMatchRight(t *testing.T) {
	right := matchTestNodes(TestTestTree.right, Sample[8+5:], false)
	if len(right) != 1 {
		t.Errorf("expecting one match, got %v", len(right))
	}
	if right[0].followUp != 0 {
		t.Errorf("expecting 0, got %v", right[0].followUp)
	}
}
