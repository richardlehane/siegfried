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

	"github.com/richardlehane/match/dwac"
	"github.com/richardlehane/siegfried/internal/siegreader"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

// identify function - brings a new matcher into existence
func (b *Matcher) identify(buf *siegreader.Buffer, quit chan struct{}, r chan core.Result, hints ...core.Hint) {
	buf.Quit = quit
	waitSet := b.priorities.WaitSet(hints...)
	maxBOF, maxEOF := b.maxBOF, b.maxEOF
	if len(hints) > 0 {
		var hasExclude bool
		for _, h := range hints {
			if h.Pivot == nil {
				hasExclude = true
				break
			}
		}
		if hasExclude {
			maxBOF, maxEOF = waitSet.MaxOffsets()
		}
	}
	incoming, resume := b.scorer(buf, waitSet, quit, r)
	rdr := siegreader.LimitReaderFrom(buf, maxBOF)
	// First test BOF frameset
	bfchan := b.bofFrames.index(buf, false, quit)
	for bf := range bfchan {
		if config.Debug() {
			fmt.Fprintln(config.Out(), strike{b.bofFrames.testTreeIndex[bf.idx], 0, bf.off, bf.length, false, true})
		}
		incoming <- strike{b.bofFrames.testTreeIndex[bf.idx], 0, bf.off, bf.length, false, true}
	}
	select {
	case <-quit: // the matcher has called quit
		close(incoming)
		return
	default:
	}
	// start bof matcher if not yet started
	b.bmu.Do(func() {
		b.bAho = dwac.New(b.bofSeq.set)
	})
	var resuming bool
	var bofOffset int64
	// Do an initial check of BOF sequences (until resume signal)
	bchan, rchan := b.bAho.Index(rdr)
	for br := range bchan {
		if br.Index[0] == -1 { // if we got a resume signal
			resuming = true
			bofOffset = br.Offset
			break
		} else {
			if config.Debug() {
				fmt.Fprintln(config.Out(), strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false})
			}
			incoming <- strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false}
		}
	}
	select {
	case <-quit: // the matcher has called quit
		close(rchan)
		close(incoming)
		return
	default:
	}
	// check the EOF
	if maxEOF != 0 {
		_, _ = buf.CanSeek(0, true) // force a full read to enable EOF scan to proceed for streams
		// EOF frame tests (should be none)
		efchan := b.eofFrames.index(buf, true, quit)
		for ef := range efchan {
			if config.Debug() {
				fmt.Fprintln(config.Out(), strike{b.eofFrames.testTreeIndex[ef.idx], 0, ef.off, ef.length, true, true})
			}
			incoming <- strike{b.eofFrames.testTreeIndex[ef.idx], 0, ef.off, ef.length, true, true}
		}
		// EOF sequences
		b.emu.Do(func() {
			b.eAho = dwac.New(b.eofSeq.set)
		})
		rrdr := siegreader.LimitReverseReaderFrom(buf, maxEOF)
		echan, erchan := b.eAho.Index(rrdr) // todo: handle the possibility of wild EOF segments
		// Scan complete EOF
		for er := range echan {
			if er.Index[0] == -1 { // handle EOF wilds (should be none!)
				incoming <- strike{-1, -1, er.Offset, 0, true, false} // send resume signal
				kfids := <-resume
				dynSet := b.eofSeq.indexes(filterTests(b.tests, kfids))
				erchan <- dynSet
				continue
			}
			if config.Debug() {
				fmt.Fprintln(config.Out(), strike{b.eofSeq.testTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false})
			}
			incoming <- strike{b.eofSeq.testTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false}
		}
		select {
		case <-quit: // the matcher has called quit
			close(rchan)
			close(incoming)
			return
		default:
		}
	}
	if !resuming {
		close(incoming)
		return
	}
	// Finally, finish BOF scan looking for wilds only
	incoming <- strike{-1, -1, bofOffset, 0, false, false} // send resume signal
	kfids := <-resume
	dynSet := b.bofSeq.indexes(filterTests(b.tests, kfids))
	rchan <- dynSet
	for br := range bchan {
		if config.Debug() {
			fmt.Fprintln(config.Out(), strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false})
		}
		incoming <- strike{b.bofSeq.testTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false}
	}
	close(incoming)
}
