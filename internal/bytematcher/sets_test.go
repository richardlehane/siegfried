package bytematcher

import (
	"testing"

	"github.com/richardlehane/match/dwac"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
)

var TestSeqSetBof = &seqSet{
	set:           []dwac.Seq{},
	testTreeIndex: []int{},
}

var TestSeqSetEof = &seqSet{
	set:           []dwac.Seq{},
	testTreeIndex: []int{},
}

var TestFrameSetBof = &frameSet{
	set:           []frames.Frame{},
	testTreeIndex: []int{},
}

func TestSeqSet(t *testing.T) {
	s := &seqSet{}
	c1 := dwac.Seq{MaxOffsets: []int64{0}, Choices: []dwac.Choice{{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c2 := dwac.Seq{MaxOffsets: []int64{0}, Choices: []dwac.Choice{{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c3 := dwac.Seq{MaxOffsets: []int64{-1}, Choices: []dwac.Choice{{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c4 := dwac.Seq{MaxOffsets: []int64{-1}, Choices: []dwac.Choice{{[]byte{'a', 'p', 'p', 'l', 'e', 's'}}}}
	s.add(c1, 0)
	i := s.add(c2, 1)
	if i != 0 {
		t.Error("Adding identical byte sequences should return a single TestTree index")
	}
	i = s.add(c3, 1)
	if i != 1 {
		t.Error("A different offset, should mean a different TestTree index")
	}
	i = s.add(c4, 2)
	if i != 2 {
		t.Error("A different choice slice, should mean a different TestTree index")
	}
	i = s.add(c2, 3)
	if i != 0 {
		t.Error("Adding identical byte sequences should return a single TestTree index")
	}
}

func TestFrameSet(t *testing.T) {
	f := &frameSet{}
	f.add(tests.TestFrames[0], 0)
	i := f.add(tests.TestFrames[0], 1)
	if i != 0 {
		t.Error("Adding identical frame sequences should return a single TestTree index")
	}
}
