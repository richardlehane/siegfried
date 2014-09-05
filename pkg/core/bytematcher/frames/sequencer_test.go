package frames_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames_common"
)

func TestSequencer(t *testing.T) {
	sequencer := NewSequencer(false)
	byts := sequencer(TestFrames[0])
	if len(byts) != 1 {
		t.Error("Sequencer: expected only one sequence")
	}
	if len(byts[0]) != 4 {
		t.Error("Sequencer: expected an initial sequence length of 5")
	}
	byts = sequencer(TestFrames[2])
	if len(byts) != 1 {
		t.Error("Sequencer: expected only one sequence")
	}
	if len(byts[0]) != 9 {
		t.Error("Sequencer: expected a final sequence length of 9")
	}
}
