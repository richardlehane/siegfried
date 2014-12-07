package process

import (
	"testing"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
)

var TestProcessObj = &Process{
	KeyFrames: [][]keyFrame{},
	Tests:     []*testTree{},
	BOFFrames: nil,
	EOFFrames: nil,
	BOFSeq:    nil,
	EOFSeq:    nil,
}

var Sample = []byte("testTESTMATCHAAAAAAAAAAAYNESStesty")

func TestProcess(t *testing.T) {
	p := New()
	p.Options = Options{8192, 2059, 9, 1}
	for i, v := range tests.TestSignatures {
		err := p.AddSignature(v)
		if err != nil {
			t.Errorf("Unexpected error adding signature; sig %v; error %v", i, v)
		}
	}
	if len(p.KeyFrames) != 6 {
		t.Errorf("Expecting 6 keyframe slices, got %d", len(p.KeyFrames))
	}
	var tl int
	for _, v := range p.KeyFrames {
		tl += len(v)
	}
	if tl != 12 {
		t.Errorf("Expecting a total of 12 keyframes, got %d", tl)
	}
	if len(p.Tests) != 9 {
		t.Errorf("Expecting a total of 9 tests, got %d", len(p.Tests))
	}
	if len(p.BOFSeq.Set) != 4 {
		t.Errorf("Expecting 4 BOF seqs, got %d", len(p.BOFSeq.Set))
	}
	e1 := wac.Seq{[]int{0}, []wac.Choice{wac.Choice{[]byte{'t', 'e', 's', 't'}}}}
	if !seqEquals(p.BOFSeq.Set[0], e1) {
		t.Errorf("Expecting %v to equal %v", p.BOFSeq.Set[0], e1)
	}
	e2 := wac.Seq{[]int{-1}, []wac.Choice{wac.Choice{[]byte{'t', 'e', 's', 't'}}}}
	if seqEquals(p.BOFSeq.Set[0], e2) {
		t.Errorf("Not expecting %v to equal %v", p.BOFSeq.Set[0], e2)
	}
	if len(p.EOFSeq.Set) != 2 {
		t.Errorf("Expecting two EOF seqs, got %d, first is %v", len(p.EOFSeq.Set), p.EOFSeq.Set[0])
	}
	if len(p.BOFFrames.Set) != 1 {
		t.Errorf("Expecting one BOF Frame, got %d", len(p.BOFFrames.Set))
	}
	if len(p.EOFFrames.Set) != 0 {
		t.Errorf("Expecting no EOF frame, got %d", len(p.EOFFrames.Set))
	}
}

func TestProcessFmt418(t *testing.T) {
	p := New()
	p.Options = Options{2000, 500, 10, 1}
	p.AddSignature(tests.TestFmts[418])
	if len(p.KeyFrames[0]) != 2 {
		for _, v := range p.KeyFrames[0] {
			t.Errorf("%s\n", v)
		}
	}
}

var test418 = "%!PS-Adobe-2.0UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU%%DocumentNeededResources:UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU%%+ procset Adobe_Illustrator_AI3"

func TestProcessFmt134(t *testing.T) {
	p := New()
	p.Options = Options{1000, 500, 3, 1}
	p.AddSignature(tests.TestFmts[134])
	if len(p.KeyFrames[0]) != 8 {
		for _, v := range p.KeyFrames[0] {
			t.Errorf("%s\n", v)
		}
	}
}
