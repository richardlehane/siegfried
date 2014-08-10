package namematcher

import (
	"bytes"
	"testing"
)

func TestIO(t *testing.T) {
	em := NewExtensionMatcher()
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
