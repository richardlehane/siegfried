package frames_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/internal/bytematcher/frames"
	. "github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
	. "github.com/richardlehane/siegfried/internal/bytematcher/patterns/tests"
)

func TestFixed(t *testing.T) {
	f2 := NewFrame(BOF, TestSequences[0], 0, 0)
	f3 := NewFrame(BOF, TestSequences[0], 0)
	if !TestFrames[0].Equals(f2) {
		t.Error("Fixed fail: Equality")
	}
	if TestFrames[0].Equals(f3) {
		t.Error("Fixed fail: Equality")
	}
	if !TestFrames[0].Equals(TestFrames[1]) {
		t.Error("Fixed fail: Equality")
	}
	num, rem, _ := f2.MaxMatches(10)
	if num != 1 {
		t.Errorf("Fixed fail: MaxMatches should have one match, got %d", num)
	}
	if rem != 6 {
		t.Errorf("Fixed fail: MaxMatches should have rem value 6, got %d", rem)
	}
}

func TestWindow(t *testing.T) {
	w2 := NewFrame(BOF, TestSequences[0], 0, 5)
	w3 := NewFrame(BOF, TestSequences[0], 0)
	if !TestFrames[5].Equals(w2) {
		t.Error("Window fail: Equality")
	}
	if TestFrames[5].Equals(w3) {
		t.Error("Window fail: Equality")
	}
	num, rem, _ := w2.MaxMatches(16)
	if num != 2 {
		t.Errorf("Window fail: MaxMatches should have two matches, got %d", num)
	}
	if rem != 12 {
		t.Errorf("Window fail: MaxMatches should have rem value 12, got %d", rem)
	}
}

func TestWild(t *testing.T) {
	w2 := NewFrame(BOF, TestSequences[0])
	w3 := NewFrame(BOF, TestSequences[0], 1)
	if !TestFrames[9].Equals(w2) {
		t.Error("Wild fail: Equality")
	}
	if TestFrames[9].Equals(w3) {
		t.Error("Wild fail: Equality")
	}
	num, rem, _ := w2.MaxMatches(10)
	if num != 3 {
		t.Errorf("Wild fail: MaxMatches should have three matches, got %d", num)
	}
	if rem != 6 {
		t.Errorf("Wild fail: MaxMatches should have rem value 6, got %d", rem)
	}
}

func TestWildMin(t *testing.T) {
	w2 := NewFrame(BOF, TestSequences[0], 5)
	w3 := NewFrame(BOF, TestSequences[0], 0, 5)
	if !TestFrames[11].Equals(w2) {
		t.Error("Wild fail: Equality")
	}
	if TestFrames[11].Equals(w3) {
		t.Error("Wild fail: Equality")
	}
	num, rem, _ := w2.MaxMatches(10)
	if num != 1 {
		t.Errorf("WildMin fail: MaxMatches should have one matches, got %d", num)
	}
	if rem != 1 {
		t.Errorf("WildMin fail: MaxMatches should have rem value 1, got %d", rem)
	}
}
