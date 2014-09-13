package bytematcher

import (
	//"fmt"
	"io"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// Identify function - brings a new matcher into existence
func (b *ByteMatcher) identify(buf *siegreader.Buffer, quit chan struct{}, r chan Result, wait chan []int) {
	buf.SetQuit(quit)
	bprog, eprog := make(chan int), make(chan int)
	gate := make(chan struct{})
	incoming := b.newMatcher(buf, quit, r, bprog, eprog, wait, gate)

	// Test BOF/EOF sequences
	var rdr io.ByteReader
	if b.MaxBOF > 0 {
		rdr = buf.NewLimitReader(b.MaxBOF)
	} else {
		rdr = buf.NewReader()
	}
	bchan := b.bAho.Index(rdr, bprog, quit)
	// Do an initial check of BOF sequences
Loop:
	for {
		select {
		case br, ok := <-bchan:
			if !ok {
				select {
				case <-quit:
					// the matcher has called quit
					close(incoming)
					return
				default:
					//	we've reached the EOF but haven't got a final match
					break Loop
				}
			} else {
				//fmt.Println(strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final})
				incoming <- strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}
			}
		case <-gate:
			break Loop
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
	} else {
		rrdr, err = buf.NewReverseReader()
	}
	if err != nil {
		close(incoming)
		return
	}
	echan := b.eAho.Index(rrdr, eprog, quit)
	for {
		select {
		case bf, ok := <-bfchan:
			if !ok {
				bfchan = nil
			} else {
				incoming <- strike{b.BOFFrames.TestTreeIndex[bf.Idx], 0, bf.Off, bf.Length, false, true, true}
			}
		case ef, ok := <-efchan:
			if !ok {
				efchan = nil
			} else {
				incoming <- strike{b.EOFFrames.TestTreeIndex[ef.Idx], 0, ef.Off, ef.Length, true, true, true}
			}
		case br, ok := <-bchan:
			if !ok {
				bchan = nil
			} else {
				//fmt.Println(strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}) // testing
				incoming <- strike{b.BOFSeq.TestTreeIndex[br.Index[0]], br.Index[1], br.Offset, br.Length, false, false, br.Final}
			}
		case er, ok := <-echan:
			if !ok {
				echan = nil
			} else {
				//fmt.Println(strike{b.EOFSeq.TestTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false, er.Final}) // testing
				incoming <- strike{b.EOFSeq.TestTreeIndex[er.Index[0]], er.Index[1], er.Offset, er.Length, true, false, er.Final}
			}
		}
		if bfchan == nil && efchan == nil && bchan == nil && echan == nil {
			close(incoming)
			return
		}
	}
}
