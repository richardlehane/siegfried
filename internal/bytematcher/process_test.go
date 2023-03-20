package bytematcher

import (
	"sync"
	"testing"

	"github.com/richardlehane/match/dwac"
	"github.com/richardlehane/siegfried/internal/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/priority"
	"github.com/richardlehane/siegfried/pkg/config"
)

var TestProcessObj = &Matcher{
	keyFrames: [][]keyFrame{},
	tests:     []*testTree{},
	bofFrames: nil,
	eofFrames: nil,
	bofSeq:    nil,
	eofSeq:    nil,
}

var Sample = []byte("testTESTMATCHAAAAAAAAAAAYNESStesty")

func newMatcher() *Matcher {
	return &Matcher{
		bofFrames:  &frameSet{},
		eofFrames:  &frameSet{},
		bofSeq:     &seqSet{},
		eofSeq:     &seqSet{},
		priorities: &priority.Set{},
		bmu:        &sync.Once{},
		emu:        &sync.Once{},
	}
}

func TestProcess(t *testing.T) {
	b := newMatcher()
	config.SetDistance(8192)()
	config.SetRange(2059)()
	config.SetChoices(9)()
	for i, v := range tests.TestSignatures {
		err := b.addSignature(v)
		if err != nil {
			t.Errorf("Unexpected error adding persist; sig %v; error %v", i, v)
		}
	}
	saver := persist.NewLoadSaver(nil)
	Save(b, saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	b = Load(loader).(*Matcher)
	if len(b.keyFrames) != 8 {
		t.Errorf("Expecting 8 keyframe slices, got %d", len(b.keyFrames))
	}
	var tl int
	for _, v := range b.keyFrames {
		tl += len(v)
	}
	if tl != 16 {
		t.Errorf("Expecting a total of 16 keyframes, got %d", tl)
	}
	if len(b.tests) != 12 {
		t.Errorf("Expecting a total of 12 tests, got %d", len(b.tests))
	}
	if len(b.bofSeq.set) != 5 {
		t.Errorf("Expecting 5 BOF seqs, got %d", len(b.bofSeq.set))
	}
	e1 := dwac.Seq{MaxOffsets: []int64{0}, Choices: []dwac.Choice{{[]byte{'t', 'e', 's', 't'}}}}
	if !seqEquals(b.bofSeq.set[0], e1) {
		t.Errorf("Expecting %v to equal %v", b.bofSeq.set[0], e1)
	}
	e2 := dwac.Seq{MaxOffsets: []int64{-1}, Choices: []dwac.Choice{{[]byte{'t', 'e', 's', 't'}}}}
	if seqEquals(b.bofSeq.set[0], e2) {
		t.Errorf("Not expecting %v to equal %v", b.bofSeq.set[0], e2)
	}
	if len(b.eofSeq.set) != 3 {
		t.Errorf("Expecting 3 EOF seqs, got %d, first is %v", len(b.eofSeq.set), b.eofSeq.set[0])
	}
	if len(b.bofFrames.set) != 1 {
		t.Errorf("Expecting one BOF Frame, got %d", len(b.bofFrames.set))
	}
	if len(b.eofFrames.set) != 0 {
		t.Errorf("Expecting no EOF frame, got %d", len(b.eofFrames.set))
	}
}

func TestProcessFmt418(t *testing.T) {
	b := newMatcher()
	config.SetDistance(2000)()
	config.SetRange(500)()
	config.SetChoices(10)()
	b.addSignature(tests.TestFmts[418])
	saver := persist.NewLoadSaver(nil)
	Save(b, saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	b = Load(loader).(*Matcher)
	if len(b.keyFrames[0]) != 2 {
		t.Errorf("Expecting 2, got %d", len(b.keyFrames[0]))
	}
}

func TestProcessFmt134(t *testing.T) {
	b := newMatcher()
	config.SetDistance(1000)
	config.SetRange(500)
	config.SetChoices(3)
	b.addSignature(tests.TestFmts[134])
	saver := persist.NewLoadSaver(nil)
	Save(b, saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	b = Load(loader).(*Matcher)
	if len(b.keyFrames[0]) != 1 {
		t.Errorf("Expecting 1, got %d", len(b.keyFrames[0]))
	}
	for _, t := range b.tests {
		t.maxLeftDistance = maxLength(t.left)
		t.maxRightDistance = maxLength(t.right)
	}
	if len(b.tests) != 1 {
		t.Errorf("Expecting 1 test, got %d", len(b.tests))
	}
}

func TestProcessFmt363(t *testing.T) {
	b := newMatcher()
	b.addSignature(tests.TestFmts[363])
	saver := persist.NewLoadSaver(nil)
	Save(b, saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	b = Load(loader).(*Matcher)
	if len(b.keyFrames[0]) != 2 {
		for _, v := range b.keyFrames[0] {
			t.Errorf("%s\n", v)
		}
	}
}
