package pronom

import (
// "testing"

// "github.com/richardlehane/siegfried/internal/bytematcher/patterns"
)

// GITHUB ACTIONS TODO.
/*
func TestRange(t *testing.T) {
	rng := Range{[]byte{1}, []byte{3}}
	rng2 := Range{[]byte{1}, []byte{3}}
	rng3 := Range{[]byte{11, 250}, []byte{12, 1}}
	if !rng.Equals(rng2) {
		t.Error("Range fail: Equality")
	}
	if r, _ := rng.Test([]byte{1}); len(r) != 1 || r[0] != 1 {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{2}); len(r) != 1 || r[0] != 1 {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{3}); len(r) != 1 || r[0] != 1 {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{4}); len(r) > 0 {
		t.Error("Range fail: Test should fail")
	}
	if r, _ := rng3.Test([]byte{251, 11}); len(r) != 1 || r[0] != 1 {
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
	if r, _ := rng.Test([]byte{1}); len(r) != 1 || r[0] != 1 {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{2}); len(r) != 1 || r[0] != 1 {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{3}); len(r) != 1 || r[0] != 1 {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{4}); len(r) > 0 {
		t.Error("Not Range fail: Test shoud fail")
	}
	if rng.NumSequences() != 253 {
		t.Error("Not Range fail: NumSequences; expecting 253 got", rng.NumSequences())
	}
	seqs := rng.Sequences()
	if len(seqs) != 253 {
		t.Error("Not Range fail: Sequences")
	}
}
*/
