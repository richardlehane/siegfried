package bytematcher

import (
	"bytes"
	"sync"
	"testing"
)

func TestMatch(t *testing.T) {
	var wg sync.WaitGroup
	matcherStub.wg = &wg
	wg.Add(1)
	err := matcherStub.buf.SetSource(bytes.NewBuffer(mStub))
	if err != nil {
		t.Errorf("matcher fail: error setting siegreader source")
	}
	go matcherStub.match(0, 10, 5, false)
	i := <-matcherStub.r
	if i != 1 {
		t.Errorf("matcher fail: expecting 1, got %d", i)
	}
	wg.Wait()
}
