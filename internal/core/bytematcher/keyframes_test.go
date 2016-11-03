package bytematcher

import (
	"testing"

	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames/tests"
)

var TestKeyFrames = []keyFrame{
	{
		typ: frames.BOF,
		seg: keyFramePos{
			pMin: 8,
			pMax: 12,
		},
	},
	{
		typ: frames.PREV,
		seg: keyFramePos{
			pMin: 5,
			pMax: 5,
		},
	},
	{
		typ: frames.PREV,
		seg: keyFramePos{
			pMin: 0,
			pMax: -1,
		},
	},
	{
		typ: frames.SUCC,
		seg: keyFramePos{
			pMin: 5,
			pMax: 10,
		},
	},
	{
		typ: frames.EOF,
		seg: keyFramePos{
			pMin: 0,
			pMax: 0,
		},
	},
}

func TestKeyFrame(t *testing.T) {
	_, left, right := toKeyFrame(tests.TestSignatures[1], position{1, 1, 2})
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
