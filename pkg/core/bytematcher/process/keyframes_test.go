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

func TestMustExist(t *testing.T) {
	if TestKeyFrames[0].MustExist(5, false) {
		t.Error("KeyFrame: BOF KeyFrame does not have to exist at offset 5")
	}
	if TestKeyFrames[0].MustExist(10, false) {
		t.Error("KeyFrame: BOF KeyFrame does not have to exist at offset 10")
	}
	if !TestKeyFrames[0].MustExist(15, false) {
		t.Error("KeyFrame: BOF KeyFrame must exist at offset 15")
	}
}
