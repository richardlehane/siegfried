package bytematcher

import (
	"testing"

	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames/tests"
)

// [BOF 0:test], [P 10-20:TESTY|YNESS], [S *:test|testy], [S 0:testy], [E 10-20:test|testy]
func TestSignatureOne(t *testing.T) {
	s := tests.TestSignatures[0].Segment(8192, 2059)
	if len(s) != 3 {
		t.Errorf("Segment fail: expecting 3 segments, got %d", len(s))
	}
	// [BOF 0:test], [P 10-20:TESTY|YNESS]
	if len(s[0]) != 2 {
		t.Errorf("Segment fail: expecting the first segment to have two frames, got %d", len(s[0]))
	}
	if characterise(s[0]) != bofZero {
		t.Errorf("Characterise fail: expecting the first segment to be bofzero, it is %v", characterise(s[0]))
	}
	pos := position{4, 0, 1}
	if bofLength(s[0], 64) != pos {
		t.Errorf("bofLength fail: expecting position %v, to equal %v", bofLength(s[0], 64), pos)
	}
	// [S *:test|testy]
	if len(s[1]) != 1 {
		t.Errorf("Segment fail: expecting the second segment to have a single frame, got %d", len(s[0]))
	}
	if characterise(s[1]) != succ {
		t.Errorf("Characterise fail: expecting the second segment to be succ, it is %v", characterise(s[1]))
	}
	// the length in varLength reports the minimum, not the maximum length
	pos = position{4, 0, 1}
	if varLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", varLength(s[1], 64), pos)
	}
	// [S 0:testy], [E 10-20:test|testy]
	if len(s[2]) != 2 {
		t.Errorf("Segment fail: expecting the last segment to have two frames, got %d", len(s[2]))
	}
	if characterise(s[2]) != eofWindow {
		t.Errorf("Characterise fail: expecting the last segment to be eofWindow, it is %v", characterise(s[2]))
	}
	pos = position{9, 0, 2}
	if varLength(s[2], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", varLength(s[2], 64), pos)
	}
}

// [BOF 0:test], [P 10-20:TESTY|YNESS], [P 0-1:TEST], [S 0:testy], [S *:test|testy], [E 0:23]
func TestSignatureTwo(t *testing.T) {
	s := tests.TestSignatures[1].Segment(8192, 2059)
	if len(s) != 3 {
		t.Errorf("Segment fail: expecting 3 segments, got %d", len(s))
	}
	// [BOF 0:test], [P 10-20:TESTY|YNESS], [P 0-1:TEST]
	if len(s[0]) != 3 {
		t.Errorf("Segment fail: expecting the first segment to have three frames, got %d", len(s[0]))
	}
	if characterise(s[0]) != bofZero {
		t.Errorf("Characterise fail: expecting the first segment to be bofzero, it is %v", characterise(s[0]))
	}
	pos := position{4, 0, 1}
	if bofLength(s[0], 64) != pos {
		t.Errorf("bofLength fail: expecting position %v, to equal %v", bofLength(s[0], 64), pos)
	}
	// [S 0:testy], [S *:test|testy]
	if len(s[1]) != 2 {
		t.Errorf("Segment fail: expecting the second segment to have two frames, got %d", len(s[1]))
	}
	if characterise(s[1]) != succ {
		t.Errorf("Characterise fail: expecting the second segment to be succ, it is %v", characterise(s[1]))
	}
	pos = position{9, 0, 2}
	if varLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", bofLength(s[1], 64), pos)
	}
}

// [BOF 0-5:a|b|c...|j], [P *:test]
func TestSignatureThree(t *testing.T) {
	s := tests.TestSignatures[2].Segment(8192, 2059)
	if len(s) != 2 {
		t.Errorf("Segment fail: expecting 2 segments, got %d", len(s))
	}
	// [BOF 0-5:a|b]
	if characterise(s[0]) != bofWindow {
		t.Errorf("Characterise fail: expecting the first segment to be bofWindow, it is %v", characterise(s[0]))
	}
	pos := position{1, 0, 1}
	if varLength(s[0], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", varLength(s[0], 64), pos)
	}
	// [P *:test]
	if len(s[1]) != 1 {
		t.Errorf("Segment fail: expecting the second segment to have one frame, got %d", len(s[1]))
	}
	if characterise(s[1]) != prev {
		t.Errorf("Characterise fail: expecting the second segment to be prev, it is %v", characterise(s[1]))
	}
	pos = position{4, 0, 1}
	if varLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", varLength(s[1], 64), pos)
	}
}

// [BOF 0:test], [P 10-20:TESTY|YNESS], [BOF *:test]
func TestSignatureFour(t *testing.T) {
	s := tests.TestSignatures[3].Segment(8192, 2059)
	if len(s) != 2 {
		t.Errorf("Segment fail: expecting 2 segments, got %d", len(s))
	}
	// [BOF 0:test], [P 10-20:TESTY|YNESS]
	if characterise(s[0]) != bofZero {
		t.Errorf("Characterise fail: expecting the first segment to be bofWindow, it is %v", characterise(s[0]))
	}
	pos := position{4, 0, 1}
	if bofLength(s[0], 64) != pos {
		t.Errorf("bofLength fail: expecting position %v, to equal %v", bofLength(s[0], 64), pos)
	}
	// [BOF *:test]
	if len(s[1]) != 1 {
		t.Errorf("Segment fail: expecting the second segment to have one frame, got %d", len(s[1]))
	}
	if characterise(s[1]) != bofWild {
		t.Errorf("Characterise fail: expecting the second segment to be prev, it is %v", characterise(s[1]))
	}
	pos = position{4, 0, 1}
	if varLength(s[1], 64) != pos {
		t.Errorf("varLength fail: expecting position %v, to equal %v", varLength(s[1], 64), pos)
	}
}

func TestFmt418(t *testing.T) {
	s := tests.TestFmts[418].Segment(2000, 500)
	if len(s) != 2 {
		t.Errorf("fmt418 fail: expecting 2 segments, got %d", len(s))
	}
	if characterise(s[0]) != bofZero {
		t.Errorf("fmt418 fail: expecting the first segment to be bofzero, got %v", characterise(s[0]))
	}
	pos := position{14, 0, 1}
	if bofLength(s[0], 2) != pos {
		t.Errorf("fmt418 fail: expecting the first segment to have pos %v, got %v", pos, bofLength(s[0], 2))
	}
	if characterise(s[1]) != prev {
		t.Errorf("fmt418 fail: expecting the second segment to be prev, got %v", characterise(s[1]))
	}
	pos = position{33, 0, 2}
	if varLength(s[1], 2) != pos {
		t.Errorf("fmt418 fail: expecting the second segment to have pos %v, got %v", pos, bofLength(s[1], 2))
		t.Error(s[1])
	}
}
