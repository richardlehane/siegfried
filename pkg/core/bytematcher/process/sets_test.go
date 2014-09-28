package process

import (
	"testing"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
)

var TestSeqSetBof = &seqSet{
	Set:           []wac.Seq{},
	TestTreeIndex: []int{},
}

var TestSeqSetEof = &seqSet{
	Set:           []wac.Seq{},
	TestTreeIndex: []int{},
}

var TestFrameSetBof = &frameSet{
	Set:           []frames.Frame{},
	TestTreeIndex: []int{},
}

func TestseqSet(t *testing.T) {
	s := &seqSet{}
	c1 := wac.Seq{[]int{0}, []wac.Choice{wac.Choice{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c2 := wac.Seq{[]int{0}, []wac.Choice{wac.Choice{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c3 := wac.Seq{[]int{-1}, []wac.Choice{wac.Choice{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c4 := wac.Seq{[]int{-1}, []wac.Choice{wac.Choice{[]byte{'a', 'p', 'p', 'l', 'e', 's'}}}}
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
