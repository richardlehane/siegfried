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
	r := bytes.NewBuffer([]byte("012345678"))
	bufs := siegreader.New()
	b, err := bufs.Get(r)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	res, _ := testMatcher.Identify("example.tt", b)
	var collect []core.Result
	for r := range res {
		collect = append(collect, r)
	}
	expect := count * 2
	if len(collect) != expect {
		t.Errorf("Expecting %d results, got %d", expect, len(collect))
		for _, r := range collect {
			t.Error(r.Basis())
		}
	}
}
