package bytematcher

import (
	"sort"
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/process"
)

type tally struct {
	*matcher
	results chan Result
	quit    chan struct{}
	wait    chan []int

	once     *sync.Once
	bofQueue *sync.WaitGroup
	eofQueue *sync.WaitGroup
	stop     chan struct{}

	bofOff int
	eofOff int

	waitList []int
	waitM    *sync.RWMutex

	kfHits chan kfHit
	halt   chan bool
}

func newTally(r chan Result, q chan struct{}, w chan []int, m *matcher) *tally {
	t := &tally{
		matcher:  m,
		results:  r,
		quit:     q,
		wait:     w,
		once:     &sync.Once{},
		bofQueue: &sync.WaitGroup{},
		eofQueue: &sync.WaitGroup{},
		stop:     make(chan struct{}),
		waitList: nil,
		waitM:    &sync.RWMutex{},
		kfHits:   make(chan kfHit),
		halt:     make(chan bool),
	}
	go t.filterHits()
	return t
}

func (t *tally) shutdown(eof bool) {
	go t.once.Do(func() { t.finalise(eof) })
}

func (t *tally) finalise(eof bool) {
	if eof {
		t.bofQueue.Wait()
		t.eofQueue.Wait()
	}
	close(t.quit)
	t.drain()
	if !eof {
		t.bofQueue.Wait()
		t.eofQueue.Wait()
	}
	close(t.results)
	close(t.stop)
}

func (t *tally) drain() {
	for {
		select {
		case _, ok := <-t.incoming:
			if !ok {
				t.incoming = nil
			}
		case _ = <-t.bofProgress:
		case _ = <-t.eofProgress:
		}
		if t.incoming == nil {
			return
		}
	}
}

type kfHit struct {
	id     process.KeyFrameID
	offset int
	length int
}

func (t *tally) filterHits() {
	var satisfied bool
	for {
		select {
		case <-t.stop:
			return
		case hit := <-t.kfHits:
			if satisfied {
				// the halt channel tells the matcher to continuing checking complete/incomplete tests for the strike
				t.halt <- true
				continue
			}
			// in case of a race
			if !t.checkWait(hit.id[0]) {
				t.halt <- false
				continue
			}
			success, basis := t.applyKeyFrame(hit.id, hit.offset, hit.length)
			if success {
				if h := t.sendResult(hit.id[0], basis); h {
					t.halt <- true
					satisfied = true
					t.shutdown(false)
					continue
				}
			}
			t.halt <- false
		}
	}
}

func (t *tally) sendResult(idx int, basis string) bool {
	t.results <- Result{idx, basis}
	w := <-t.wait // every result sent must result in a new priority list being returned & we need to drain this or it will block
	// nothing more to wait for
	if len(w) == 0 {
		return true
	}
	t.setWait(w)
	return false
}

func (t *tally) setWait(w []int) {
	t.waitM.Lock()
	t.waitList = w
	t.waitM.Unlock()
}

// check a signature ID against the priority list
func (t *tally) checkWait(i int) bool {
	t.waitM.RLock()
	defer t.waitM.RUnlock()
	if t.waitList == nil {
		return true
	}
	idx := sort.SearchInts(t.waitList, i)
	if idx == len(t.waitList) || t.waitList[idx] != i {
		return false
	}
	return true
}

// check to see whether should still wait for signatures in the priority list, given the offset
// trim the wait list if possible
// An issue is the buffering we do with wacs: how can we know that an incoming strike isn't just in transit, even though earlier than the reported offset?
// Perhaps change WAC so progress is set as a special Result{progress: true}, that way they can still be buffered but are in order
// This would simplify the initial Identify loop: rather than waiting for gate to close, can just listen for progress results in that loop
func (t *tally) continueWait(o int) bool {
	t.waitM.Lock()
	defer t.waitM.Unlock()
	if t.waitList == nil {
		// if we don't have a wait list, we've got no matches and must continue
		return true
	}
	// check the strikecache ??
	return true
}
