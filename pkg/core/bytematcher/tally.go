// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bytematcher

import (
	"sync"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/process"
	"github.com/richardlehane/siegfried/pkg/core/priority"
)

type tally struct {
	*matcher
	results chan core.Result
	quit    chan struct{}

	once     *sync.Once
	bofQueue *sync.WaitGroup
	eofQueue *sync.WaitGroup
	stop     chan struct{}

	waitSet *priority.WaitSet

	kfHits chan kfHit
	halt   chan bool
}

func newTally(r chan core.Result, q chan struct{}, m *matcher) *tally {
	t := &tally{
		matcher:  m,
		results:  r,
		quit:     q,
		once:     &sync.Once{},
		bofQueue: &sync.WaitGroup{},
		eofQueue: &sync.WaitGroup{},
		stop:     make(chan struct{}),
		waitSet:  m.bm.Priorities.WaitSet(),
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
	// drain any remaining matches
	for _ = range t.incoming {
	}
	if !eof {
		t.bofQueue.Wait()
		t.eofQueue.Wait()
	}
	close(t.results)
	close(t.stop)
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
			if !t.waitSet.Check(hit.id[0]) {
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
	return t.waitSet.Put(idx)
}

// check to see whether should still wait for signatures in the priority list, given the offset
func (t *tally) continueWait(o int, rev bool) bool {
	w := t.waitSet.WaitingOn()
	// must continue if any of the waitlists are nil
	if w == nil {
		return true
	}
	for _, v := range w {
		kf := t.bm.KeyFrames[v]
		if rev {
			for i := len(kf) - 1; i >= 0 && kf[i].Typ > frames.PREV; i-- {
				if kf[i].Key.PMax == -1 || kf[i].Key.PMax+kf[i].Key.LMax > o {
					return true
				}
			}
		} else {
			for _, f := range kf {
				if f.Typ > frames.PREV {
					break
				}
				if f.Key.PMax == -1 || f.Key.PMax+f.Key.LMax > o {
					return true
				}
			}
		}
	}
	return false
}
