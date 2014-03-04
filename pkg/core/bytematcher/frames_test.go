package bytematcher

import "testing"

func TestFixed(t *testing.T) {
	f2 := NewFrame(BOF, seqStub2, 0, 0)
	f3 := NewFrame(BOF, seqStub2, 0)
	if !fixedStub.Equals(f2) {
		t.Error("Fixed fail: Equality")
	}
	if fixedStub.Equals(f3) {
		t.Error("Fixed fail: Equality")
	}
	if !fixedStub3.Equals(fixedStub3) {
		t.Error("Fixed fail: Equality")
	}
}

func TestWindow(t *testing.T) {
	w2 := NewFrame(BOF, seqStub2, 0, 5)
	w3 := NewFrame(BOF, seqStub2, 0)
	if !windowStub.Equals(w2) {
		t.Error("Window fail: Equality")
	}
	if windowStub.Equals(w3) {
		t.Error("Window fail: Equality")
	}
}

func TestWild(t *testing.T) {
	w2 := NewFrame(BOF, seqStub2)
	w3 := NewFrame(BOF, seqStub2, 1)
	if !wildStub.Equals(w2) {
		t.Error("Wild fail: Equality")
	}
	if wildStub.Equals(w3) {
		t.Error("Wild fail: Equality")
	}
}

func TestWildMin(t *testing.T) {
	w2 := NewFrame(BOF, seqStub2, 5)
	w3 := NewFrame(BOF, seqStub2, 0, 5)
	if !wildMinStub.Equals(w2) {
		t.Error("Wild fail: Equality")
	}
	if wildMinStub.Equals(w3) {
		t.Error("Wild fail: Equality")
	}
}
