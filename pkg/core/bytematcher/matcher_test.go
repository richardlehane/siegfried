package bytematcher

import (
	"testing"
	"time"
)

func TestMatch(t *testing.T) {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	go matcherStub.match(0, 10, 5, false)
	select {
	case i := <-matcherStub.r:
		if i != 1 {
			t.Errorf("matcher fail: expecting 1, got %d", i)
		}
	case <-timeout:
		t.Errorf("matcher fail: timeout")
	}
}
