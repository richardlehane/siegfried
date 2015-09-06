package bytematcher

import (
	"bytes"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames/tests"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

func dsetup() (chan<- strike, <-chan core.Result) {
	bm := New()
	bm.Add(SignatureSet(tests.TestSignatures), nil)
	bufs := siegreader.New()
	buf, _ := bufs.Get(bytes.NewBuffer(TestSample1))
	buf.SizeNow()
	res := make(chan core.Result)
	return bm.demon(buf, make(chan struct{}), res), res
}

func TestDemon(t *testing.T) {
	scorer, res := dsetup()
	scorer <- strike{0, 0, 0, 4, false, false, true}
	scorer <- strike{1, 0, 17, 9, true, false, false}
	scorer <- strike{1, 1, 30, 5, true, false, true}
	if r := <-res; r.Index() != 0 {
		t.Errorf("expecting result %d, got %d", 0, r.Index())
	}
}

// 5 Sept 15 BenchmarkScorer	   20000	   117391 ns/op
func BenchmarkDemon(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		bench.StopTimer()
		scorer, res := setup()
		bench.StartTimer()
		scorer <- strike{0, 0, 0, 4, false, false, true}
		scorer <- strike{1, 0, 17, 9, true, false, false}
		scorer <- strike{1, 1, 30, 5, true, false, true}
		_ = <-res
	}
}
