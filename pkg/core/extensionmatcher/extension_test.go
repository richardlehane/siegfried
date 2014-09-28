package extensionmatcher

import (
	"bytes"
	"testing"
)

var fmts = []string{"wav", "doc", "xls", "pdf", "ppt"}

var em = New()

func init() {
	for i, v := range fmts {
		em.Add(v, i)
	}
}

func TestWavMatch(t *testing.T) {
	res := em.Identify("hello/apple.wav", nil)
	e := <-res
	if e.Index() != 0 {
		t.Errorf("Expecting 0, got %v", e)
	}
	e, ok := <-res
	if ok {
		t.Error("Expecting a length of 1")
	}
}

func TestNoMatch(t *testing.T) {
	res := em.Identify("hello/apple.tty", nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestNoExt(t *testing.T) {
	res := em.Identify("hello/apple", nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestIO(t *testing.T) {
	em := New()
	em.Add(".doc", 1)
	em.Add(".ppt", 2)
	str := em.String()
	buf := &bytes.Buffer{}
	sz, err := em.Save(buf)
	if err != nil {
		t.Error(err)
	}
	if sz < 10 {
		t.Errorf("Save extension matcher: too small, only got %v", sz)
	}
	newem, err := Load(buf)
	if err != nil {
		t.Error(err)
	}
	str2 := newem.String()
	if str != str2 {
		t.Errorf("Load extension matcher: expecting first extension matcher (%v), to equal second extension matcher (%v)", str, str2)
	}
}
