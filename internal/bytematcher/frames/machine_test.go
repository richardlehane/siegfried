package frames_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/internal/bytematcher/frames"
	. "github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
)

func TestMachine(t *testing.T) {
	machine := Machine(TestFmts[13405])
	if machine.NumSequences() != 0 {
		t.Errorf("Expecting 0 sequences, got %d", machine.NumSequences())
	}
	// test BOF matching
	ok, l := machine.Test(TestMP3)
	if !ok {
		t.Error("Expecting the machine to match the MP3")
	}
	if l != 5218 {
		t.Errorf("Expecting length of the match to be 5218, got %d", l)
	}
	// check for pernicious slowdown
	nok, _ := machine.Test(TestBumper)
	if nok {
		t.Error("Expecting the machine not to match bumper")
	}
	// test EOF matching
	rmachine := Machine(TestFmts[13401])
	if rmachine.NumSequences() != 0 {
		t.Errorf("Expecting 0 sequences, got %d", machine.NumSequences())
	}
	ok, l = rmachine.TestR(TestMP3)
	if !ok {
		t.Error("Expecting the machine to match the MP3")
	}
	if l != 5218 {
		t.Errorf("Expecting length of the match to be 5218, got %d", l)
	}
}
