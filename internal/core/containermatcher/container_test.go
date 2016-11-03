package containermatcher

import (
	"testing"

	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/internal/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/internal/core/persist"
	"github.com/richardlehane/siegfried/internal/core/priority"
	"github.com/richardlehane/siegfried/internal/core/siegreader"
)

func testTrigger([]byte) bool {
	return true
}

var testContainerMatcher *ContainerMatcher = &ContainerMatcher{
	ctype:        ctype{testTrigger, newTestReader},
	conType:      0,
	nameCTest:    make(map[string]*cTest),
	priorities:   &priority.Set{},
	startIndexes: []int{0},
	entryBufs:    siegreader.New(),
}

var count int

func TestMatcher(t *testing.T) {
	ctypes = []ctype{{testTrigger, newTestReader}}
	// test adding
	count++
	testMatcher, _, err := Add(Matcher{testContainerMatcher},
		SignatureSet{
			0,
			[][]string{{"one", "two"}, {"one"}},
			[][]frames.Signature{{tests.TestSignatures[3], tests.TestSignatures[4]}, {tests.TestSignatures[2]}},
		},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	// test IO
	str := testMatcher.String()
	saver := persist.NewLoadSaver(nil)
	Save(testMatcher, saver)
	if len(saver.Bytes()) < 100 {
		t.Errorf("Save container: too small, only got %v", len(saver.Bytes()))
	}
	newcm := Load(persist.NewLoadSaver(saver.Bytes()))
	str2 := newcm.String()
	if len(str) != len(str2) {
		t.Errorf("Load container: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}
