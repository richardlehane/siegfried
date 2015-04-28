package bytematcher

// TODO: something!

/*
import (
	"bytes"
	"testing"
)


// Matcher
var TestMatcher *matcher = &matcher{
	b:                BmStub,
	buf:              siegreader.New(),
	r:                make(chan int),
	partialKeyframes: make(map[[2]int][][2]int),
	limit:            nil,
	limitm:           &sync.RWMutex{},
	limitc:           nil,
	incoming:         make(chan strike),
	quit:             make(chan struct{}),
}

func TestMatch(t *testing.T) {
	err := matcherStub.buf.SetSource(bytes.NewBuffer(mStub))
	if err != nil {
		t.Errorf("matcher fail: error setting siegreader source")
	}

	go matcherStub.match()
	for {
		select {
		case matcherStub.incoming <- strike{0, 10, 5, false, false}:
		case i := <-matcherStub.r:
			if i != 1 {
				t.Errorf("matcher fail: expecting 1, got %d", i)
			}
			return
		}
	}
}
*/
