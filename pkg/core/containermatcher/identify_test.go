package containermatcher

import (
	"bytes"
	"io"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

func TestIdentify(t *testing.T) {
	ctypes = append(ctypes, ctype{testTrigger, newTestReader})
	// test adding
	err := testContainerMatcher.addSignature([]string{"one", "two"}, []frames.Signature{tests.TestSignatures[3], tests.TestSignatures[4]})
	if err != nil {
		t.Fatal(err)
	}
	err = testContainerMatcher.addSignature([]string{"one"}, []frames.Signature{tests.TestSignatures[4]})
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
	r := bytes.NewBuffer([]byte("012345678"))
	bufs := siegreader.New()
	b, err := bufs.Get(r)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	res := testMatcher.Identify("example.tt", b)
	var collect []core.Result
	for r := range res {
		collect = append(collect, r)
	}
	if len(collect) != 2 {
		t.Errorf("Expecting 2 results, got %v", len(collect))
		for _, r := range collect {
			t.Error(r.Basis())
		}
	}
}
