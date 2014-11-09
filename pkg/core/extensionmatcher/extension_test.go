package extensionmatcher

import (
	"bytes"
	"testing"
)

var fmts = SignatureSet{[]string{"wav"}, []string{"doc"}, []string{"xls"}, []string{"pdf"}, []string{"ppt"}}

var em = New()

func init() {
	em.Add(fmts, nil)
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
	em.Add(SignatureSet{[]string{".bla"}, []string{".doc"}, []string{".ppt"}}, nil)
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
