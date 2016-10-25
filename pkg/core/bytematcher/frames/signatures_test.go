package frames_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
)

func TestContains(t *testing.T) {
	if !TestSignatures[0].Contains(TestSignatures[0]) {
		t.Error("Contains: expecting identical signatures to be contained")
	}
}
