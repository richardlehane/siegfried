package namematcher

import "testing"

var fmts = []string{"wav", "doc", "xls", "pdf", "ppt"}

var em = NewExtensionMatcher()

func init() {
	for i, v := range fmts {
		em.Add(v, i)
	}
}

func TestWavMatch(t *testing.T) {
	res := em.Identify("hello/apple.wav")
	if len(res) != 1 {
		t.Fatalf("Expecting a length of 1, got %v", len(res))
	}

	if res[0] != 0 {
		t.Errorf("Expecting 0, got %v", res[0])
	}
}

func TestNoMatch(t *testing.T) {
	res := em.Identify("hello/apple.tty")
	if len(res) > 0 {
		t.Errorf("Should not match, got %v", len(res))
	}
}

func TestNoExt(t *testing.T) {
	res := em.Identify("hello/apple")
	if len(res) > 0 {
		t.Errorf("Should not match, got %v", len(res))
	}
}
