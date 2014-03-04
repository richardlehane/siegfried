package bytematcher

import "testing"

func TestSegment(t *testing.T) {
	s := segment(sigStub, 10, 10)
	if len(s) != 4 {
		t.Errorf("Segment fail: expecting 4 signatures, got %d", len(s))
	}
}
