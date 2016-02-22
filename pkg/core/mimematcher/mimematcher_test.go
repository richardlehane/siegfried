package mimematcher

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/persist"
)

var fmts = SignatureSet{[]string{"application/json"}, []string{"application/json;v1"}, []string{"text/plain"}, []string{"x-world/x-3dmf"}, []string{"application/x-cocoa"}}

var sm = New()

func init() {
	sm.Add(fmts, nil)
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
	sm := New()
	sm.Add(fmts, nil)
	str := sm.String()
	saver := persist.NewLoadSaver(nil)
	sm.Save(saver)
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
