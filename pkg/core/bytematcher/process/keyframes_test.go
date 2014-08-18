package process

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

func TestKeyFrame(t *testing.T) {
	_, left, right := toKeyFrame(frames.TestSignatures[1], position{1, 1, 2})
	if len(left) != 1 {
		t.Error("KeyFrame: expecting only one frame on the left")
	}
	seq := left[0].Pat().Sequences()
	if seq[0][1] != 'e' {
		t.Error("KeyFrame: expecting the left frame's pattern to have been reversed")
	}
	if len(right) != 4 {
		t.Errorf("KeyFrame: expecting three frames on the right, got %d", len(right))
	}
}
