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
	"fmt"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

func (b *Matcher) start(bof bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if bof {
		if b.bAho != nil {
			return
		}
		if b.lowmem {
			b.bAho = wac.NewLowMem(b.bofSeq.set)
			return
		}
		b.bAho = wac.New(b.bofSeq.set)
		return
	}
	if b.eAho != nil {
		return
	}
	if b.lowmem {
		b.eAho = wac.NewLowMem(b.eofSeq.set)
		return
	}
	b.eAho = wac.New(b.eofSeq.set)
}

// Identify function - brings a new matcher into existence
func (b *Matcher) identify(buf siegreader.Buffer, quit chan struct{}, r chan core.Result) {
	buf.SetQuit(quit)
	incoming := b.newScorer(buf, quit, r)
	rdr := siegreader.LimitReaderFrom(buf, b.maxBOF)

	// First test BOF frameset
	bfchan := b.bofFrames.index(buf, false, quit)
	for bf := range bfchan {
		if config.Debug() {
			fmt.Println(strike{b.bofFrames.testTreeIndex[bf.idx], 0, bf.off, bf.length, false, true, true})
		}
		incoming <- strike{b.bofFrames.testTreeIndex[bf.idx], 0, bf.off, bf.length, false, true, true}
	}
	select {
	case <-quit: // the matcher has called quit
		close(incoming)
		return
	default:
	}

	// Do an initial check of BOF sequences
	b.start(true) // start bof matcher if not yet started
	var bchan chan wac.Result
	bchan = b.bAho.Index(rdr)
	/*for br := range bchan {
		if br.Index[0] == -1 {
			incoming <- progressStrike(br.Offset, false)
			if br.Offset > 2048 {
				break
			}
		} else {
			if config.Debug() {
				fmt.Println(strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final})
			}
			incoming <- strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}
		}
	}
	select {
	case <-quit: // the matcher has called quit
		for _ = range bchan {
		} // drain first
		close(incoming)
		return
	default:
	}*/

	// Check EOF frame tests
	efchan := b.eofFrames.index(buf, true, quit)
	for ef := range efchan {
		if config.Debug() {
			fmt.Println(strike{b.eofFrames.testTreeIndex[ef.idx], 0, ef.off, ef.length, true, true, true})
		}
		incoming <- strike{b.eofFrames.testTreeIndex[ef.idx], 0, ef.off, ef.length, true, true, true}
	}

	// Setup EOF sequences test
	b.start(false)
	rrdr := siegreader.LimitReverseReaderFrom(buf, b.maxEOF)
	echan := b.eAho.Index(rrdr)

	// Now enter main search loop
	for {
		select {
		case br, ok := <-bchan:
			if !ok {
				bchan = nil
			} else {
				if br.Index[0] == -1 {
					incoming <- progressStrike(br.Offset, false)
				} else {
					if config.Debug() {
						fmt.Println(strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final})
					}
					incoming <- strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}
				}
			}
		case er, ok := <-echan:
			if !ok {
				echan = nil
			} else {
				if er.Index[0] == -1 {
					incoming <- progressStrike(er.Offset, true)
				} else {
					if config.Debug() {
						fmt.Println(strike{b.eofSeq.testTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false, er.Final})
					}
					incoming <- strike{b.eofSeq.testTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false, er.Final}
				}
			}
		}
		if bchan == nil && echan == nil {
			close(incoming)
			return
		}
	}
}
