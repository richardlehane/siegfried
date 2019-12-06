package patterns_test

import (
	"testing"

	. "github.com/richardlehane/siegfried/internal/bytematcher/patterns"
	. "github.com/richardlehane/siegfried/internal/bytematcher/patterns/tests"
)

func TestBMH(t *testing.T) {
	b := NewBMHSequence(TestSequences[0])
	b1 := NewBMHSequence(TestSequences[0])
	if !b.Equals(b1) {
		t.Error("BMH equality fail")
	}
	ok, l := b.Test([]byte("test"))
	if ok < 0 {
		t.Error("Expecting bmh to match test")
	}
	if ok != 4 {
		t.Errorf("Expecting bmh match length to be 4, got %d", l)
	}
	ok, l = b.Test([]byte("tost"))
	if ok > -1 {
		t.Error("Not expecting bmh to match tost")
	}
	if l != 3 {
		t.Errorf("Expecting bmh skip to be 3, got %d", l)
	}
}

func TestRBMH(t *testing.T) {
	b := NewRBMHSequence(TestSequences[0])
	ok, l := b.TestR([]byte("tosttest"))
	if ok < 0 {
		t.Error("Expecting bmh to match test")
	}
	if ok != 4 {
		t.Errorf("Expecting bmh match length to be 4, got %d", l)
	}
	ok, l = b.TestR([]byte("testtost"))
	if ok > -1 {
		t.Error("Not expecting bmh to match tost")
	}
	if l != 3 {
		t.Errorf("Expecting bmh skip to be 3, got %d", l)
	}
}
