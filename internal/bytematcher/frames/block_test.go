package frames_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/internal/bytematcher/frames"
	. "github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
)

func TestBlock(t *testing.T) {
	segs := Signature(TestSignatures[7]).Segment(1000, 500, 500, 5)
	if len(segs) != 3 {
		t.Fatalf("Expecting three frames after running segment, got %d", len(segs))
	}
	blocks := Blockify(segs[0])
	if len(blocks) != 1 {
		t.Fatalf("Expecting one frame after running blockify, got %d, %v", len(blocks), blocks)
	}
	blk, ok := blocks[0].Pattern.(*Block)
	if !ok {
		t.Fatal("The pattern should be a block!")
	}
	hits, ju := blk.Test([]byte("test01234test"))
	if len(hits) != 1 || hits[0] != 13 || ju != 3 {
		t.Errorf("Expecting a single hit, length 13, with a jump of 3; got %v and %d", hits, ju)
	}
	blocks = Blockify(segs[1])
	if _, ok := blocks[0].Pattern.(*Block); ok {
		t.Fatal("Second segment should not be a block!")
	}
	blocks = Blockify(segs[2])
	if len(blocks) != 1 {
		t.Fatalf("Expecting one frame after running blockify, got %d", len(blocks))
	}
	blk, ok = blocks[0].Pattern.(*Block)
	if !ok {
		t.Fatalf("Last segment should be a block!")
	}
	hits, ju = blk.TestR([]byte("testy23"))
	if len(hits) != 1 || hits[0] != 7 || ju != 5 {
		t.Errorf("Expecting a single hit, length 7, with a jump of 5; got %v and %d", hits, ju)
	}
}
