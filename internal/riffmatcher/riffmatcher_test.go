package riffmatcher

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/core"
)

var testdata = flag.String("testdata", filepath.Join("..", "..", "cmd", "sf", "testdata"), "override the default test data directory")

var fmts = SignatureSet{
	[4]byte{'a', 'f', 's', 'p'},
	[4]byte{'I', 'C', 'R', 'D'},
	[4]byte{'I', 'S', 'F', 'T'},
	[4]byte{'I', 'C', 'M', 'T'},
	[4]byte{'f', 'a', 'c', 't'},
	[4]byte{'f', 'm', 't', ' '},
	[4]byte{'d', 'a', 't', 'a'},
	[4]byte{'I', 'N', 'F', 'O'},
	[4]byte{'W', 'A', 'V', 'E'},
}

var rm core.Matcher

func init() {
	rm, _, _ = Add(rm, fmts, nil)
}

func TestMatch(t *testing.T) {
	f, err := os.Open(filepath.Join(*testdata, "benchmark", "Benchmark.wav"))
	if err != nil {
		t.Fatal(err)
	}
	bufs := siegreader.New()
	b, _ := bufs.Get(f)
	res, err := rm.Identify("", b)
	if err != nil {
		t.Fatal(err)
	}
	var hits []int
	for h := range res {
		hits = append(hits, h.Index())
	}
	if len(hits) != len(fmts) {
		t.Fatalf("Expecting %d hits, got %d", len(fmts), len(hits))
	}
}

func TestIO(t *testing.T) {
	str := rm.String()
	saver := persist.NewLoadSaver(nil)
	Save(rm, saver)
	if len(saver.Bytes()) < 10 {
		t.Errorf("Save riff matcher: too small, only got %v", saver.Bytes())
	}
	loader := persist.NewLoadSaver(saver.Bytes())
	newrm := Load(loader)
	str2 := newrm.String()
	if str != str2 {
		t.Errorf("Load riff matcher: expecting first matcher (%v), to equal second matcher (%v)", str, str2)
	}
}
