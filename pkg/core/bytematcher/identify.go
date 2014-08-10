package bytematcher

import (
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// Identify function - brings a new matcher into existence
func (b *ByteMatcher) identify(buf *siegreader.Buffer, quit chan struct{}, r chan int, wait chan []int) {
	buf.SetQuit(quit)
	bprog, eprog := make(chan int), make(chan int)
	m := b.newMatcher(buf, quit, r, bprog, eprog, wait)
	go m.match()
	// Test BOF/EOF frames
	bfchan := b.BOFFrames.Index(buf, false, quit)
	efchan := b.EOFFrames.Index(buf, true, quit)
	// Test BOF/EOF sequences
	bchan := b.bAho.Index(buf.NewReader(), bprog, quit)
	rrdr, err := buf.NewReverseReader()
	if err != nil {
		// not much in the way of error returns at this stage!
		close(m.incoming)
		return
	}
	echan := b.eAho.Index(rrdr, eprog, quit)
	for {
		select {
		case bf, ok := <-bfchan:
			if !ok {
				bfchan = nil
			} else {
				m.incoming <- strike{bf.Idx, 0, bf.Off, bf.Length, false, true, true}
			}
		case ef, ok := <-efchan:
			if !ok {
				efchan = nil
			} else {
				m.incoming <- strike{ef.Idx, 0, ef.Off, ef.Length, true, true, true}
			}
		case br, ok := <-bchan:
			if !ok {
				bchan = nil
			} else {
				m.incoming <- strike{br.Index[0], br.Index[1], br.Offset, br.Length, false, false, br.Final}
			}
		case er, ok := <-echan:
			if !ok {
				echan = nil
			} else {
				m.incoming <- strike{er.Index[0], er.Index[1], er.Offset, er.Length, true, false, er.Final}
			}
		}
		if bfchan == nil && efchan == nil && bchan == nil && echan == nil {
			close(m.incoming)
			break
		}
	}
}
