package process

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

func TestSegment(t *testing.T) {
	p := New()
	p.SetOptions(10, 10)
	s := p.splitSegments(frames.TestSignatures[0])
	if len(s) != 4 {
		t.Errorf("Segment fail: expecting 4 signatures, got %d", len(s))
	}
}
