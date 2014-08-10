package bytematcher

import (
	"bytes"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

func TestNew(t *testing.T) {
	New()
}

func TestIO(t *testing.T) {
	bm, err := Signatures(frames.TestSignatures)
	if err != nil {
		t.Error(err)
	}
	str := bm.String()
	buf := &bytes.Buffer{}
	sz, err := bm.Save(buf)
	if err != nil {
		t.Error(err)
	}
	if sz < 100 {
		t.Errorf("Save bytematcher: too small, only got %v", sz)
	}
	newbm, err := Load(buf)
	if err != nil {
		t.Error(err)
	}
	str2 := newbm.String()
	if str != str2 {
		t.Errorf("Load bytematcher: expecting first bytematcher (%v), to equal second bytematcher (%v)", str, str2)
	}
}
