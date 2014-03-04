package pronom

import (
	"testing"
)

func TestNotSequence(t *testing.T) {
	seq := NotSequence{'t', 'e', 's', 't'}
	seq2 := NotSequence{'t', 'e', 's', 't'}
	seq3 := NotSequence{255}
	seq4 := NotSequence{0}
	seq5 := NotSequence{10}
	if !seq.Equals(seq2) {
		t.Error("NotSequence fail: Equality test")
	}
	if r, _ := seq.Test([]byte{'t', 'e', 's', 't'}); r {
		t.Error("NotSequence fail: Test")
	}
	if _, l := seq.Test([]byte{'t', 'o', 'o', 't'}); l != 4 {
		t.Error("NotSequence fail: Test")
	}
	seqs := seq3.Sequences()
	if len(seqs) != 255 {
		t.Error("NotSequence fail: Sequences")
	}
	seqs = seq4.Sequences()
	if len(seqs) != 255 {
		t.Error("NotSequence fail: Sequences")
	}
	seqs = seq5.Sequences()
	if len(seqs) != 255 {
		t.Error("NotSequence fail: Sequences")
	}
}

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
	rng := NotRange{[]byte{1}, []byte{3}}
	rng2 := NotRange{[]byte{1}, []byte{3}}
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
