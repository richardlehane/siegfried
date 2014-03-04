package namematcher

import "testing"

var fmts = []string{"wav", "doc", "xls", "pdf", "ppt"}

var em = NewExtensionMatcher()

func TestWavMatch(t *testing.T) {
	for i, v := range fmts {
		em.Add(v, i)
	}
	em.SetName("hello/apple.wav")
	res := em.Match()
	if i := <-res; i != 0 {
		t.Errorf("Expecting 0, got %v", i)
	}
}

func TestNoMatch(t *testing.T) {
	em.SetName("hello/apple.tty")
	res := em.Match()
	for r := range res {
		t.Errorf("Should not match, got %v", r)
	}
}

func TestNoExt(t *testing.T) {
	em.SetName("hello/apple")
	res := em.Match()
	for r := range res {
		t.Errorf("Should not match, got %v", r)
	}
}
