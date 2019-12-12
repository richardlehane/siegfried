package frames_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/internal/bytematcher/frames"
	. "github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
)

func TestContains(t *testing.T) {
	if !TestSignatures[0].Contains(TestSignatures[0]) {
		t.Error("Contains: expecting identical signatures to be contained")
	}
}

func TestMirror(t *testing.T) {
	mirror := TestSignatures[2].Mirror()
	if len(mirror) < 2 || mirror[1].Orientation() != EOF {
		t.Errorf("Mirror fail: got %v", mirror)
	}
}
