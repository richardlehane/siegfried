package bytematcher

import (
	"testing"

	wac "github.com/richardlehane/match/fwac"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/persist"
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

func TestProcess(t *testing.T) {
	m, _, _ := Add(nil, SignatureSet{}, nil)
	b := m.(*Matcher)
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
	if len(b.keyFrames) != 6 {
		t.Errorf("Expecting 6 keyframe slices, got %d", len(b.keyFrames))
	}
	var tl int
	for _, v := range b.keyFrames {
		tl += len(v)
	}
	if tl != 12 {
		t.Errorf("Expecting a total of 12 keyframes, got %d", tl)
	}
	if len(b.tests) != 9 {
		t.Errorf("Expecting a total of 9 tests, got %d", len(b.tests))
	}
	if len(b.bofSeq.set) != 4 {
		t.Errorf("Expecting 4 BOF seqs, got %d", len(b.bofSeq.set))
	}
	e1 := wac.Seq{[]int64{0}, []wac.Choice{wac.Choice{[]byte{'t', 'e', 's', 't'}}}}
	if !seqEquals(b.bofSeq.set[0], e1) {
		t.Errorf("Expecting %v to equal %v", b.bofSeq.set[0], e1)
	}
	e2 := wac.Seq{[]int64{-1}, []wac.Choice{wac.Choice{[]byte{'t', 'e', 's', 't'}}}}
	if seqEquals(b.bofSeq.set[0], e2) {
		t.Errorf("Not expecting %v to equal %v", b.bofSeq.set[0], e2)
	}
	if len(b.eofSeq.set) != 2 {
		t.Errorf("Expecting two EOF seqs, got %d, first is %v", len(b.eofSeq.set), b.eofSeq.set[0])
	}
	if len(b.bofFrames.set) != 1 {
		t.Errorf("Expecting one BOF Frame, got %d", len(b.bofFrames.set))
	}
	if len(b.eofFrames.set) != 0 {
		t.Errorf("Expecting no EOF frame, got %d", len(b.eofFrames.set))
	}
}

func TestProcessFmt418(t *testing.T) {
	m, _, _ := Add(nil, SignatureSet{}, nil)
	b := m.(*Matcher)
	config.SetDistance(2000)()
	config.SetRange(500)()
	config.SetChoices(10)()
	b.addSignature(tests.TestFmts[418])
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

var test418 = "%!PS-Adobe-2.0UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU%%DocumentNeededResources:UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU%%+ procset Adobe_Illustrator_AI3"

func TestProcessFmt134(t *testing.T) {
	m, _, _ := Add(nil, SignatureSet{}, nil)
	b := m.(*Matcher)
	config.SetDistance(1000)
	config.SetRange(500)
	config.SetChoices(3)
	b.addSignature(tests.TestFmts[134])
	saver := persist.NewLoadSaver(nil)
	Save(b, saver)
	loader := persist.NewLoadSaver(saver.Bytes())
	b = Load(loader).(*Matcher)
	if len(b.keyFrames[0]) != 8 {
		for _, v := range b.keyFrames[0] {
			t.Errorf("%s\n", v)
		}
	}
	for _, t := range b.tests {
		t.maxLeftDistance = maxLength(t.left)
		t.maxRightDistance = maxLength(t.right)
	}
	if len(b.tests) != 8 {
		for _, v := range b.tests {
			t.Error(v.maxRightDistance)
			t.Error(v.right)
		}
	}
}

func TestProcessFmt363(t *testing.T) {
	m, _, _ := Add(nil, SignatureSet{}, nil)
	b := m.(*Matcher)
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
