package containermatcher

import (
	"bytes"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

func testTrigger([]byte) bool {
	return true
}

var testContainerMatcher *ContainerMatcher = &ContainerMatcher{
	ctype:      ctype{testTrigger, newTestReader},
	CType:      2,
	NameCTest:  make(map[string]*CTest),
	Priorities: &priority.Set{},
	Sindexes:   []int{0},
}

var testMatcher Matcher = []*ContainerMatcher{testContainerMatcher}

func TestMatcher(t *testing.T) {
	ctypes = append(ctypes, ctype{testTrigger, newTestReader})
	// test adding
	err := testContainerMatcher.addSignature([]string{"one", "two"}, []frames.Signature{tests.TestSignatures[3], tests.TestSignatures[4]})
	if err != nil {
		t.Fatal(err)
	}
	err = testContainerMatcher.addSignature([]string{"one"}, []frames.Signature{tests.TestSignatures[2]})
	if err != nil {
		t.Fatal(err)
	}
	testContainerMatcher.Priorities.Add(nil, 2)
	for _, v := range testContainerMatcher.NameCTest {
		err = v.commit(nil, 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	// test IO
	str := testMatcher.String()
	buf := &bytes.Buffer{}
	sz, err := testMatcher.Save(buf)
	if err != nil {
		t.Error(err)
	}
	if sz < 100 {
		t.Errorf("Save container: too small, only got %v", sz)
	}
	newcm, err := Load(buf)
	if err != nil {
		t.Error(err)
	}
	str2 := newcm.String()
	if len(str) != len(str2) {
		t.Errorf("Load container: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}
