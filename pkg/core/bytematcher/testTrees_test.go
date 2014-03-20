package bytematcher

import "testing"

func TestMaxLength(t *testing.T) {
	tStub := newTestTree()
	tStub.add([2]int{0, 0}, []Frame{}, []Frame{fixedStub, fixedStub4, windowStub4})
	tStub.add([2]int{0, 0}, []Frame{}, []Frame{fixedStub2, windowStub2})
	if maxLength(tStub.Right) != 29 {
		t.Errorf("maxLength fail: expecting 29 got %v", maxLength(tStub.Right))
	}
}

func TestMatchLeft(t *testing.T) {
	left := matchTestNodes(testTreeStub.Left, mStub[:10], true)
	if len(left) != 1 {
		t.Errorf("expecting one match, got %v", len(left))
	}
	if left[0].followUp != 0 {
		t.Errorf("expecting 0, got %v", left[0].followUp)
	}
}

func TestMatchRight(t *testing.T) {
	right := matchTestNodes(testTreeStub.Right, mStub[10+5:], false)
	if len(right) != 1 {
		t.Errorf("expecting one match, got %v", len(right))
	}
	if right[0].followUp != 0 {
		t.Errorf("expecting 0, got %v", right[0].followUp)
	}
}
