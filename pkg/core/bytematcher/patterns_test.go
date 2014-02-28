package bytematcher

import "testing"

func TestSequence(t *testing.T) {
	if !seqStub.Equals(seqStub2) {
		t.Error("Seq fail: Equality")
	}
	if r, _ := seqStub.Test([]byte{'t', 'o', 'o', 't'}); r {
		t.Error("Sequence fail: shouldn't match")
	}
	if _, l := seqStub.Test([]byte{'t', 'e', 's', 't', 'y'}); l != 5 {
		t.Error("Sequence fail: should match")
	}
	reverseSeq := seqStub.Reverse()
	if reverseSeq[1] != 't' || reverseSeq[2] != 's' || reverseSeq[3] != 'e' || reverseSeq[4] != 't' {
		t.Error("Sequence fail: Reverse")
	}
}

func TestChoice(t *testing.T) {
	if !choiceStub.Equals(choiceStub2) {
		t.Error("Choice fail: Equality")
	}
	if _, l := choiceStub.Test([]byte{'t', 'e', 's', 't'}); l != 4 {
		t.Error("Choice test fail: Test")
	}
	if choiceStub.NumSequences() != 2 {
		t.Error("Choice fail: NumSequences; expecting 2 got", choiceStub.NumSequences())
	}
	seqStubs := choiceStub.Sequences()
	if seqStubs[0][0] != 't' || seqStubs[1][0] != 't' {
		t.Error("Choice fail: Sequences; expecting t, t got ", seqStubs[0][0], seqStubs[1][0])
	}
}
