package frames

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
)

func TestFixed(t *testing.T) {
	f2 := NewFrame(BOF, patterns.TestSequences[0], 0, 0)
	f3 := NewFrame(BOF, patterns.TestSequences[0], 0)
	if !TestFrames[0].Equals(f2) {
		t.Error("Fixed fail: Equality")
	}
	if TestFrames[0].Equals(f3) {
		t.Error("Fixed fail: Equality")
	}
	if !TestFrames[0].Equals(TestFrames[1]) {
		t.Error("Fixed fail: Equality")
	}
}

func TestWindow(t *testing.T) {
	w2 := NewFrame(BOF, patterns.TestSequences[0], 0, 5)
	w3 := NewFrame(BOF, patterns.TestSequences[0], 0)
	if !TestFrames[5].Equals(w2) {
		t.Error("Window fail: Equality")
	}
	if TestFrames[5].Equals(w3) {
		t.Error("Window fail: Equality")
	}
}

func TestWild(t *testing.T) {
	w2 := NewFrame(BOF, patterns.TestSequences[0])
	w3 := NewFrame(BOF, patterns.TestSequences[0], 1)
	if !TestFrames[9].Equals(w2) {
		t.Error("Wild fail: Equality")
	}
	if TestFrames[9].Equals(w3) {
		t.Error("Wild fail: Equality")
	}
}

func TestWildMin(t *testing.T) {
	w2 := NewFrame(BOF, patterns.TestSequences[0], 5)
	w3 := NewFrame(BOF, patterns.TestSequences[0], 0, 5)
	if !TestFrames[11].Equals(w2) {
		t.Error("Wild fail: Equality")
	}
	if TestFrames[11].Equals(w3) {
		t.Error("Wild fail: Equality")
	}
}
