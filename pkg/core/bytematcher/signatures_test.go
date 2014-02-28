package bytematcher

import "testing"

func (abs absoluteFrame) verify(off OffType, min, max int, f Frame) bool {
	if off != abs.OffType || min != abs.min || max != abs.max {
		return false
	}
	if abs.Frame.Equals(f) {
		return true
	}
	return false
}

func TestAbsolutes(t *testing.T) {
	bAbs, eAbs := sigStub.absolutes()
	if len(bAbs) != 2 {
		t.Error("Fail: expecting two BOF frames, got ", len(bAbs))
		t.FailNow()
	}
	if !bAbs[0].verify(BOF, 0, 0, fixedStub2) {
		t.Error("Fail: first BOF absoluteFrame is incorrect, got ", bAbs[0])
	}
	if !bAbs[1].verify(BOF, 14, 24, windowStub2) {
		t.Error("Fail: second BOF absoluteFrame is incorrect, got ", bAbs[1])
	}
	if len(eAbs) != 2 {
		t.Error("Fail: expecting two EOF frames, got ", len(eAbs))
		t.FailNow()
	}
	if !eAbs[1].verify(EOF, 14, 24, fixedStub3) {
		t.Error("Fail: second EOF absoluteFrame is incorrect, got", eAbs[1])
	}
}

func TestLongest(t *testing.T) {
	l, s, e := varLength(sigStub2, 64)
	if l != 9 {
		t.Errorf("Fail: expecting 9, got %v", l)
	}
	if s != 2 {
		t.Errorf("Fail: expecting 2, got %v", s)
	}
	if e != 4 {
		t.Errorf("Fail: expecting 4, got %v", e)
	}
}
