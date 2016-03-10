package bytematcher

import (
	"bytes"
	"io"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

var TestSample1 = []byte("test12345678910YNESSjunktestyjunktestytest12345678910111223") // should match sigs 0, 1 and 2

var TestSample2 = []byte("test12345678910YNESSjTESTunktestyjunktestytest12345678910111223") // should match all 4 sigs

func TestIO(t *testing.T) {
	bm, _, err := Add(nil, SignatureSet(tests.TestSignatures), nil)
	if err != nil {
		t.Error(err)
	}
	saver := persist.NewLoadSaver(nil)
	Save(bm, saver)
	if len(saver.Bytes()) < 100 {
		t.Errorf("Save bytematcher: too small, only got %v", len(saver.Bytes()))
	}
	newbm := Load(persist.NewLoadSaver(saver.Bytes()))
	nsaver := persist.NewLoadSaver(nil)
	Save(newbm, nsaver)
	if len(nsaver.Bytes()) != len(saver.Bytes()) {
		t.Fatalf("expecting the bms to match length: %d and %d", len(saver.Bytes()), len(nsaver.Bytes()))
	}
	if string(nsaver.Bytes()) != string(saver.Bytes()) {
		t.Errorf("Load bytematcher: expecting first bytematcher (%v), to equal second bytematcher (%v)", bm.String(), newbm.String())
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
	bm, _, err := Add(nil, SignatureSet(tests.TestSignatures), nil)
	if err != nil {
		t.Error(err)
	}
	bufs := siegreader.New()
	buf, err := bufs.Get(bytes.NewBuffer(TestSample1))
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	res, _ := bm.Identify("", buf)
	results := make([]core.Result, 0)
	for i := range res {
		results = append(results, i)
	}
	if !contains(results, []int{0, 2, 3, 4}) {
		t.Errorf("Missing result, got: %v, expecting:%v\n", results, bm)
	}
	buf, err = bufs.Get(bytes.NewBuffer(TestSample2))
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	res, _ = bm.Identify("", buf)
	results = results[:0]
	for i := range res {
		results = append(results, i)
	}
	if !contains(results, []int{0, 1, 2, 3, 4}) {
		t.Errorf("Missing result, got: %v, expecting:%v\n", results, bm)
	}
}
