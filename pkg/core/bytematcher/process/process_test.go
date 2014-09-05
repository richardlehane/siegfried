package process

import (
	"testing"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames_common"
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
	p.SetOptions(8192, 2059, 9, 1)
	for i, v := range frames_common.TestSignatures {
		err := p.AddSignature(v)
		if err != nil {
			t.Errorf("Unexpected error adding signature; sig %v; error %v", i, v)
		}
	}
	if len(p.KeyFrames) != 5 {
		t.Errorf("Expecting 5 keyframe slices, got %d", len(p.KeyFrames))
	}
	var tl int
	for _, v := range p.KeyFrames {
		tl += len(v)
	}
	if tl != 11 {
		t.Errorf("Expecting a total of 11 keyframes, got %d", tl)
	}
	if len(p.Tests) != 8 {
		t.Errorf("Expecting a total of 8 tests, got %d", len(p.Tests))
	}
	if len(p.BOFSeq.Set) != 3 {
		t.Errorf("Expecting three BOF seqs, got %d", len(p.BOFSeq.Set))
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
