package frames_test

import (
	"testing"
	. "github.com/richardlehane/siegfried/internal/bytematcher/frames"
	. "github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
)

func TestBlock(t *testing.T) {
	blocks := Blockify(TestSignatures[0])
	if len(blocks) != 3 {
		t.Fatalf("Expecting three frames after running Blockify, got %d", len(blocks))
	}
	bl1, ok := blocks[0].Pattern.(Block)
	if !ok {
		t.Fatalf("First segment should be a block!")
	}
	hits, ju := bl1.Test([]byte("test0123456789testy"))
	if len(hits) != 1 || hits[0] != 19 || ju != 3 {
		t.Errorf("Expecting a single hit, length 19, with a jump of 3; got %v and %d", hits, ju)
	}
	if _, ok := blocks[1].Pattern.(Block); ok {
		t.Fatalf("Second segment should not be a block!")	
	}
	bl2, ok := blocks[2].Pattern.(Block)
	if !ok {
		t.Fatalf("Last segment should be a block!")
	}
	hits, ju := bl2.TestR([]byte("testytesty0123456789"))
	if len(hits) != 1 || hits[0] != 20 || ju != 4 {
		t.Errorf("Expecting a single hit, length 20, with a jump of 4; got %v and %d", hits, ju)
	}
}
