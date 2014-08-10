package process

import (
	"testing"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

func TestseqSet(t *testing.T) {
	s := newseqSet()
	if s == nil {
		t.Error("Failed to create  new seqSet")
	}
	c1 := wac.Seq{0, []wac.Choice{wac.Choice{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	c2 := wac.Seq{0, []wac.Choice{wac.Choice{[]byte{'a', 'p', 'p', 'l', 'e'}}}}
	s.add(c1, 0)
	i := s.add(c2, 1)
	if i != 0 {
		t.Error("Adding identical byte sequences should return a single TestTree index")
	}
}

func TestFrameSet(t *testing.T) {
	f := newFrameSet()
	if f == nil {
		t.Error("Failed to create  new seqSet")
	}
	f.add(frames.TestFrames[0], 0)
	i := f.add(frames.TestFrames[0], 1)
	if i != 0 {
		t.Error("Adding identical frame sequences should return a single TestTree index")
	}
}
