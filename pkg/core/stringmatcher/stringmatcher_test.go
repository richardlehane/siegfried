package stringmatcher

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/persist"
)

var fmts = SignatureSet{"wav", "doc", "xls", "pdf", "ppt"}

var sm core.Matcher

func init() {
	sm, _, _ = Add(sm, fmts, nil)
}

func TestWavMatch(t *testing.T) {
	res, _ := sm.Identify(NormaliseExt("hello/apple.wav"), nil)
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
	res, _ := sm.Identify(NormaliseExt("hello/apple.tty"), nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestNoExt(t *testing.T) {
	res, _ := sm.Identify(NormaliseExt("hello/apple"), nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestIO(t *testing.T) {
	sm, _, _ = Add(sm, SignatureSet{".bla", ".doc", ".ppt"}, nil)
	str := sm.String()
	saver := persist.NewLoadSaver(nil)
	Save(sm, saver)
	if len(saver.Bytes()) < 10 {
		t.Errorf("Save string matcher: too small, only got %v", saver.Bytes())
	}
	loader := persist.NewLoadSaver(saver.Bytes())
	newsm := Load(loader)
	str2 := newsm.String()
	if str != str2 {
		t.Errorf("Load string matcher: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}
