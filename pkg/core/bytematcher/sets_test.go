package bytematcher

import "testing"

func TestSeqSet(t *testing.T) {
	s := newSeqSet()
	if s == nil {
		t.Error("Failed to create  new SeqSet")
	}
	s.add([]byte{'a', 'p', 'p', 'l', 'e'}, 0)
	i := s.add([]byte{'a', 'p', 'p', 'l', 'e'}, 1)
	if i != 0 {
		t.Error("Adding identical byte sequences should return a single TestTree index")
	}
}

func TestFrameSet(t *testing.T) {
	f := newFrameSet()
	if f == nil {
		t.Error("Failed to create  new SeqSet")
	}
	f.add(fixedStub, 0)
	i := f.add(fixedStub, 1)
	if i != 0 {
		t.Error("Adding identical frame sequences should return a single TestTree index")
	}
}
