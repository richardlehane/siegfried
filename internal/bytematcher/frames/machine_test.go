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
	l, _ := machine.Test(TestMP3)
	if len(l) < 1 {
		t.Error("Expecting the machine to match the MP3")
	}
	if l[0] != 5218 {
		t.Errorf("Expecting length of the match to be 5218, got %d", l)
	}
	// check for pernicious slowdown
	l, _ = machine.Test(TestBumper)
	if len(l) > 0 {
		t.Error("Expecting the machine not to match bumper")
	}
	// test EOF matching
	rmachine := Machine(TestFmts[13401])
	if rmachine.NumSequences() != 0 {
		t.Errorf("Expecting 0 sequences, got %d", machine.NumSequences())
	}
	l, _ = rmachine.TestR(TestMP3)
	if len(l) < 1 {
		t.Error("Expecting the machine to match the MP3")
	}
	if l[0] != 5218 {
		t.Errorf("Expecting length of the match to be 5218, got %d", l)
	}
	min, max := rmachine.Length()
	if min != 344 || max != 10450 {
		t.Errorf("Got lengths %d and %d", min, max)
	}
}

func TestMultiLenMatching(t *testing.T) {
	machine := Machine(TestSignatures[6])
	l, _ := machine.Test(TestMultiLen)
	if len(l) < 1 || l[0] != 9 {
		t.Error("Expected the machine to match the multi-len string TESTYNESS")
	}
}
