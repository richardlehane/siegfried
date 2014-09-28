package bytematcher

import (
	"bytes"
	"io"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

var TestSample1 = []byte("test12345678910YNESSjunktestyjunktestytest12345678910111223") // should match sigs 0, 1 and 2

var TestSample2 = []byte("test12345678910YNESSjTESTunktestyjunktestytest12345678910111223") // should match all 4 sigs

func TestNew(t *testing.T) {
	New()
}

func TestIO(t *testing.T) {
	bm, err := Signatures(tests.TestSignatures)
	if err != nil {
		t.Error(err)
	}
	str := bm.String()
	buf := &bytes.Buffer{}
	sz, err := bm.Save(buf)
	if err != nil {
		t.Error(err)
	}
	if sz < 100 {
		t.Errorf("Save bytematcher: too small, only got %v", sz)
	}
	newbm, err := Load(buf)
	if err != nil {
		t.Error(err)
	}
	str2 := newbm.String()
	if str != str2 {
		t.Errorf("Load bytematcher: expecting first bytematcher (%v), to equal second bytematcher (%v)", str, str2)
	}
}

func contains(a []core.Result, b []int) bool {
	for _, v := range a {
		var present bool
		for _, w := range b {
			if v.Index() == w {
				present = true
			}
		}
		if !present {
			return false
		}
	}
	return true
}

func TestMatch(t *testing.T) {
	bm, err := Signatures(tests.TestSignatures, 8192, 2059, 9, 1)
	if err != nil {
		t.Error(err)
	}
	buf := siegreader.New()
	err = buf.SetSource(bytes.NewBuffer(TestSample1))
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	res := bm.Identify("", buf)
	results := make([]core.Result, 0)
	for i := range res {
		results = append(results, i)
	}
	if !contains(results, []int{0, 2, 3, 4}) {
		t.Errorf("Missing result, got: %v, expecting:%v\n", results, bm)
	}
	err = buf.SetSource(bytes.NewBuffer(TestSample2))
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	res = bm.Identify("", buf)
	results = results[:0]
	for i := range res {
		results = append(results, i)
	}
	if !contains(results, []int{0, 1, 2, 3, 4}) {
		t.Errorf("Missing result, got: %v, expecting:%v\n", results, bm)
	}
}
