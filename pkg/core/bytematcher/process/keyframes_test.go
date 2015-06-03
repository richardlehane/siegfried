package process

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
)

var TestKeyFrames = []keyFrame{
	keyFrame{
		Typ: frames.BOF,
		Seg: keyFramePos{
			PMin: 8,
			PMax: 12,
		},
	},
	keyFrame{
		Typ: frames.PREV,
		Seg: keyFramePos{
			PMin: 5,
			PMax: 5,
		},
	},
	keyFrame{
		Typ: frames.PREV,
		Seg: keyFramePos{
			PMin: 0,
			PMax: -1,
		},
	},
	keyFrame{
		Typ: frames.SUCC,
		Seg: keyFramePos{
			PMin: 5,
			PMax: 10,
		},
	},
	keyFrame{
		Typ: frames.EOF,
		Seg: keyFramePos{
			PMin: 0,
			PMax: 0,
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
