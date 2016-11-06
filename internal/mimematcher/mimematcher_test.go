package mimematcher

import (
	"testing"

	"github.com/richardlehane/siegfried/core"
	"github.com/richardlehane/siegfried/internal/persist"
)

var fmts = SignatureSet{"application/json", "application/json;v1", "text/plain", "x-world/x-3dmf", "application/x-cocoa"}

var sm core.Matcher

func init() {
	sm, _, _ = Add(nil, fmts, nil)
}

func TestJsonMatch(t *testing.T) {
	res, _ := sm.Identify("application/json;v1", nil)
	e := <-res
	if e.Index() != 1 {
		t.Errorf("Expecting 0, got %v", e)
	}
	e = <-res
	if e.Index() != 0 {
		t.Errorf("Expecting 1, got %v", e)
	}
	_, ok := <-res
	if ok {
		t.Error("Expecting a length of 2")
	}
}

func TestNoMatch(t *testing.T) {
	res, _ := sm.Identify("application/java", nil)
	_, ok := <-res
	if ok {
		t.Error("Should not match")
	}
}

func TestIO(t *testing.T) {
	str := sm.String()
	saver := persist.NewLoadSaver(nil)
	Save(sm, saver)
	if len(saver.Bytes()) < 10 {
		t.Errorf("Save mime matcher: too small, only got %v", saver.Bytes())
	}
	loader := persist.NewLoadSaver(saver.Bytes())
	newsm := Load(loader)
	str2 := newsm.String()
	if str != str2 {
		t.Errorf("Load mime matcher: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}
