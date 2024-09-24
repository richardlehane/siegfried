package pronom

import (
	"testing"

	"github.com/richardlehane/siegfried/internal/bytematcher/patterns"
)

func TestRange(t *testing.T) {
	rng := Range{[]byte{1}, []byte{3}}
	rng2 := Range{[]byte{1}, []byte{3}}
	rng3 := Range{[]byte{11, 250}, []byte{12, 1}}
	rng4 := Range{[]byte{00, 00}, []byte{10, 00}}
	if !rng.Equals(rng2) {
		t.Error("Range fail: Equality")
	}
	if r, _ := rng.Test([]byte{1}); len(r) != 1 || r[0] != 1 {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{2}); len(r) != 1 || r[0] != 1 {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{3}); len(r) != 1 && r[0] != 1 {
		t.Error("Range fail: Test")
	}
	if r, _ := rng.Test([]byte{4}); len(r) > 0 {
		t.Error("Range fail: Test should fail")
	}
	if r, _ := rng3.Test([]byte{11, 251}); len(r) != 1 && r[0] != 2 {
		t.Errorf("Range fail: Test multibyte range, got %d, %d", len(r), r[0])
	}
	if r, _ := rng4.Test([]byte{10, 00}); len(r) != 1 && r[0] != 2 {
		t.Errorf("Range fail: Test multibyte range, got %d, %d", len(r), r[0])
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
	if r, _ := rng.Test([]byte{1}); len(r) != 0 {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{2}); len(r) != 0 {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{3}); len(r) != 0 {
		t.Error("Not Range fail: Test")
	}
	if r, _ := rng.Test([]byte{4}); len(r) != 1 || r[0] != 1 {
		t.Error("Not Range fail: 4 falls outside range so should succeed")
	}
	if rng.NumSequences() != 253 {
		t.Error("Not Range fail: NumSequences; expecting 253 got", rng.NumSequences())
	}
	seqs := rng.Sequences()
	if len(seqs) != 253 {
		t.Error("Not Range fail: Sequences")
	}
}
