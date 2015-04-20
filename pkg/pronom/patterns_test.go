package pronom

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
)

func TestRange(t *testing.T) {
	rng := Range{[]byte{1}, []byte{3}}
	rng2 := Range{[]byte{1}, []byte{3}}
	rng3 := Range{[]byte{11, 250}, []byte{12, 1}}
	if !rng.Equals(rng2) {
		t.Error("Range fail: Equality")
	}
	if r, _ := rng.Test([]byte{1}); !r {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{2}); !r {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{3}); !r {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{4}); r {
		t.Error("Range fail: Test")
	}
	if _, l := rng3.Test([]byte{251, 11}); l != 1 {
		t.Error("Range fail: Test")
	}
	if rng.NumSequences() != 3 {
		t.Error("Range fail: NumSequences")
	}
	if rng3.NumSequences() != 8 {
		t.Error("Range fail: NumSequences; expecting 8 got ", rng3.NumSequences())
	}
}

func TestNotRange(t *testing.T) {
	rng := patterns.Not{Range{[]byte{1}, []byte{3}}}
	rng2 := patterns.Not{Range{[]byte{1}, []byte{3}}}
	if !rng.Equals(rng2) {
		t.Error("NotRange fail: Equality")
	}
	if r, _ := rng.Test([]byte{1}); r {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{2}); r {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{3}); r {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{4}); !r {
		t.Error("Not Range fail: Test")
	}
	if rng.NumSequences() != 253 {
		t.Error("Not Range fail: NumSequences; expecting 253 got", rng.NumSequences())
	}
	seqs := rng.Sequences()
	if len(seqs) != 253 {
		t.Error("Not Range fail: Sequences")
	}
}

func TestMask(t *testing.T) {
	mask := Mask(0xAA)
	if r, _ := mask.Test([]byte{0xEE}); !r {
		t.Errorf("mask fail: 0xEE should match")
	}
	if r, _ := mask.Test([]byte{0x0A}); r {
		t.Errorf("mask fail: expected 0x0A not to match!")
	}
	num := mask.NumSequences()
	if num != 16 {
		t.Fatal("mask fail: expecting 16 sequences")
	}
	seqs := mask.Sequences()
	if len(seqs) != 16 {
		t.Fatal("mask fail: expecting 16 sequences")
	}
	for i, v := range seqs {
		if v[0] == 0xEE {
			break
		}
		if i == len(seqs)-1 {
			t.Fatal("mask fail: expecting 0xEE amongst sequences")
		}
	}
}

func TestAnyMask(t *testing.T) {
	amask := AnyMask(0xAA)
	if r, _ := amask.Test([]byte{0x0A}); !r {
		t.Errorf("any mask fail: 0x0A should match")
	}
	if r, _ := amask.Test([]byte{5}); r {
		t.Errorf("any mask fail: expected 5 not to match!")
	}
	num := amask.NumSequences()
	if num != 240 {
		t.Fatal("any mask fail: expecting 240 sequences")
	}
	seqs := amask.Sequences()
	if len(seqs) != 240 {
		t.Fatal("any mask fail: expecting 240 sequences")
	}
	for i, v := range seqs {
		if v[0] == 0x0A {
			break
		}
		if i == len(seqs)-1 {
			t.Fatal("any mask fail: expecting 0x0A amongst sequences")
		}
	}
}
