package bytematcher

import (
	"bytes"
	"sync"
)

func (b *Bytematcher) identify(r chan int) {
	var wg sync.WaitGroup
	m := NewMatcher(b, r, b.buf.Bytes(), &wg)
	bchan := b.bAho.IndexFixed(bytes.NewReader(b.buf.Bytes()))
	vchan := b.vAho.Index(bytes.NewReader(b.buf.Bytes()))
	echan := b.eAho.IndexFixed(newReverseReader(b.buf.Bytes()))

	for i, f := range b.BofFrames.Set {
		if match, matches := f.Match(b.buf.Bytes()); match {
			min, _ := f.Length()
			for _, off := range matches {
				wg.Add(1)
				go m.match(b.BofFrames.TestTreeIndex[i], off-min, min, false)
			}
		}
	}
	for i, f := range b.EofFrames.Set {
		if match, matches := f.MatchR(b.buf.Bytes()); match {
			for _, off := range matches {
				wg.Add(1)
				go m.match(b.EofFrames.TestTreeIndex[i], off, 0, true)
			}
		}
	}
	for {
		select {
		case bi, ok := <-bchan:
			if !ok {
				bchan = nil
			} else {
				wg.Add(1)
				go m.match(b.BofSeqs.TestTreeIndex[bi], 0, len(b.BofSeqs.Set[bi]), false)
			}
		case vi, ok := <-vchan:
			if !ok {
				vchan = nil
			} else {
				wg.Add(1)
				go m.match(b.VarSeqs.TestTreeIndex[vi.Index], vi.Offset, len(b.VarSeqs.Set[vi.Index]), false)
			}
		case ei, ok := <-echan:
			if !ok {
				echan = nil
			} else {
				wg.Add(1)
				go m.match(b.EofSeqs.TestTreeIndex[ei], 0, len(b.EofSeqs.Set[ei]), true)
			}
		}
		if bchan == nil && vchan == nil && echan == nil {
			break
		}
	}
	wg.Wait()
	close(r)
}
