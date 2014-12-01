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
	"io"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// Identify function - brings a new matcher into existence
func (b *Matcher) identify(buf *siegreader.Buffer, quit chan struct{}, r chan core.Result) {
	buf.SetQuit(quit)
	incoming := b.newMatcher(buf, quit, r)

	// Test BOF/EOF sequences
	var rdr io.ByteReader
	if b.MaxBOF > 0 {
		rdr = buf.NewLimitReader(b.MaxBOF)
	} else if b.MaxBOF < 0 {
		rdr = buf.NewReader()
	}
	// start bof matcher if not yet started
	if rdr != nil && !b.bstarted {
		b.bAho = wac.New(b.BOFSeq.Set)
		b.bstarted = true
	}
	var bchan chan wac.Result
	if rdr != nil {
		bchan = b.bAho.Index(rdr, quit)

		// Do an initial check of BOF sequences
		for br := range bchan {
			if br.Index[0] == -1 {
				progressStrike.offset = br.Offset
				incoming <- progressStrike
				if br.Offset > 8192 {
					break
				}
			} else {
				if config.Debug() {
					fmt.Println(strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final})
				}
				incoming <- strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}
			}
		}
		select {
		case <-quit:
			// the matcher has called quit
			close(incoming)
			return
		default:
			//	we've reached the EOF but haven't got a final match
			break
		}
	}
	// Test BOF/EOF frames
	bfchan := b.BOFFrames.Index(buf, false, quit)
	efchan := b.EOFFrames.Index(buf, true, quit)
	// Test EOF sequences
	var rrdr io.ByteReader
	var err error
	if b.MaxEOF > 0 {
		rrdr, err = buf.NewLimitReverseReader(b.MaxEOF)
	} else if b.MaxEOF < 0 {
		rrdr, err = buf.NewReverseReader()
	}
	if err != nil {
		close(incoming)
		return
	}
	// start EOF matcher if not yet started
	if rrdr != nil && !b.estarted {
		b.eAho = wac.New(b.EOFSeq.Set)
		b.estarted = true
	}
	var echan chan wac.Result
	if rrdr != nil {
		echan = b.eAho.Index(rrdr, quit)
	}
	for {
		select {
		case bf, ok := <-bfchan:
			if !ok {
				bfchan = nil
			} else {
				if config.Debug() {
					fmt.Println(strike{b.BOFFrames.TestTreeIndex[bf.Idx], 0, bf.Off, bf.Length, false, true, true})
				}
				incoming <- strike{b.BOFFrames.TestTreeIndex[bf.Idx], 0, bf.Off, bf.Length, false, true, true}
			}
		case ef, ok := <-efchan:
			if !ok {
				efchan = nil
			} else {
				if config.Debug() {
					fmt.Println(strike{b.EOFFrames.TestTreeIndex[ef.Idx], 0, ef.Off, ef.Length, true, true, true})
				}
				incoming <- strike{b.EOFFrames.TestTreeIndex[ef.Idx], 0, ef.Off, ef.Length, true, true, true}
			}
		case br, ok := <-bchan:
			if !ok {
				bchan = nil
			} else {
				if br.Index[0] == -1 {
					progressStrike.offset = br.Offset
					incoming <- progressStrike
				} else {
					if config.Debug() {
						fmt.Println(strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final})
					}
					incoming <- strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}
				}
			}
		case er, ok := <-echan:
			if !ok {
				echan = nil
			} else {
				if er.Index[0] == -1 {
					progressStrike.offset = er.Offset
					progressStrike.reverse = true
					incoming <- progressStrike
				} else {
					if config.Debug() && er.Index[0] != -1 {
						fmt.Println(strike{b.EOFSeq.TestTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false, er.Final})
					}
					incoming <- strike{b.EOFSeq.TestTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false, er.Final}
				}
			}
		}
		if bfchan == nil && efchan == nil && bchan == nil && echan == nil {
			close(incoming)
			return
		}
	}
}
