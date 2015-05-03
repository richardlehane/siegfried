package containermatcher

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/core/signature"
)

func testTrigger([]byte) bool {
	return true
}

var testContainerMatcher *ContainerMatcher = &ContainerMatcher{
	ctype:      ctype{testTrigger, newTestReader},
	CType:      0,
	NameCTest:  make(map[string]*CTest),
	Priorities: &priority.Set{},
	Sindexes:   []int{0},
	entryBufs:  siegreader.New(),
}

var testMatcher Matcher = []*ContainerMatcher{testContainerMatcher}

var count int

func TestMatcher(t *testing.T) {
	ctypes = []ctype{ctype{testTrigger, newTestReader}}
	// test adding
	count++
	_, err := testMatcher.Add(
		SignatureSet{
			0,
			[][]string{[]string{"one", "two"}, []string{"one"}},
			[][]frames.Signature{[]frames.Signature{tests.TestSignatures[3], tests.TestSignatures[4]}, []frames.Signature{tests.TestSignatures[2]}},
		},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	// test IO
	str := testMatcher.String()
	saver := signature.NewLoadSaver(nil)
	testMatcher.Save(saver)
	if len(saver.Bytes()) < 100 {
		t.Errorf("Save container: too small, only got %v", len(saver.Bytes()))
	}
	newcm := Load(signature.NewLoadSaver(saver.Bytes()))
	str2 := newcm.String()
	if len(str) != len(str2) {
		t.Errorf("Load container: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}
