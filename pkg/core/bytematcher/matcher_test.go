package bytematcher

import (
	"sync"
	"testing"
)

func TestMatch(t *testing.T) {
	var wg sync.WaitGroup
	matcherStub.wg = &wg
	wg.Add(1)
	go matcherStub.match(0, 10, 5, false)
	i := <-matcherStub.r
	if i != 1 {
		t.Errorf("matcher fail: expecting 1, got %d", i)
	}
	wg.Wait()
}
