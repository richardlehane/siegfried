package pronom

import (
	"io"
	"sync"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type testBMatcher struct{}

func (t testBMatcher) String() string {
	return ""
}

func (t testBMatcher) Save(w io.Writer) (int, error) {
	return 0, nil
}

func (t testBMatcher) Start()                        {}
func (t testBMatcher) SetPriorities(l priority.List) {}

func (t testBMatcher) Identify(sb *siegreader.Buffer) chan bytematcher.Result {
	ret := make(chan bytematcher.Result)
	go func() {
		ret <- bytematcher.Result{1, ""}
		ret <- bytematcher.Result{2, ""}
		close(ret)
	}()
	return ret
}

type testNMatcher struct{}

func (t testNMatcher) Identify(n string) []int {
	return []int{0}
}

func (t testNMatcher) String() string {
	return ""
}

func (t testNMatcher) Save(w io.Writer) (int, error) {
	return 0, nil
}

var testIdentifier = &PronomIdentifier{
	BPuids: []string{"fmt/1", "fmt/2", "fmt/3"},
	PuidsB: map[string][]int{"fmt/1": []int{0}, "fmt/2": []int{1}, "fmt/3": []int{2}},
	EPuids: []string{"fmt/1", "fmt/2", "fmt/3"},
	bm:     testBMatcher{},
	em:     testNMatcher{},
	ids:    pids{},
}

func TestIdentify(t *testing.T) {
	c := make(chan core.Identification)
	buf := siegreader.New()
	var wg sync.WaitGroup
	wg.Add(1)
	go testIdentifier.Identify(buf, "test.doc", c, &wg)
	i := <-c
	if i.String() != "fmt/3" {
		t.Error("expecting fmt/3")
	}
	wg.Wait()
}
