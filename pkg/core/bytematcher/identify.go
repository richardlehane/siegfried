package bytematcher

import (
	//"fmt"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// Identify function - brings a new matcher into existence
func (b *ByteMatcher) identify(buf *siegreader.Buffer, quit chan struct{}, r chan int, wait chan []int) {
	buf.SetQuit(quit)
	bprog, eprog := make(chan int), make(chan int)
	incoming := b.newMatcher(buf, quit, r, bprog, eprog, wait)

	// Test BOF/EOF sequences
	bchan := b.bAho.Index(buf.NewReader(), bprog, quit)
	// Do an initial check of BOF sequences here - until first or second send on bprog
	// TODO

	// Test BOF/EOF frames
	bfchan := b.BOFFrames.Index(buf, false, quit)
	efchan := b.EOFFrames.Index(buf, true, quit)
	// Test EOF sequences
	rrdr, err := buf.NewReverseReader()
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
			break
		}
	}
}
