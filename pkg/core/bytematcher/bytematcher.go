// Package Bytematcher builds a matching engine from a set of signatures and performs concurrent matching against an input siegreader.Buffer.
package bytematcher

import (
	"fmt"
	"io"

	"github.com/richardlehane/match/ac"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"

	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

// Bytematcher structure. Clients shouldn't need to get or set these fields directly, they are only exported so that this structure can be serialised and deserialised by encoding/gob.
type ByteMatcher struct {
	Sigs [][]keyFrame

	TestSet []*testTree

	BofSeqs *seqSet
	EofSeqs *seqSet
	VarSeqs *seqSet

	BofFrames *frameSet
	EofFrames *frameSet

	bAho *ac.Ac
	eAho *ac.Ac
	vAho *ac.Ac
}

// Create a new Bytematcher from a slice of signatures.
// Can give optional distance, range, choice, variable sequence length values to override the defaults of 8192, 2048, 64.
//   - the distance and range values dictate how signatures are segmented during processing
//   - the choices value controls how signature segments are converted into simple byte sequences during processing
//   - the varlen value controls what is the minimum length sequence acceptable for the variable Aho Corasick tree. The longer this length, the fewer false matches you will get during searching.
//
// Example:
//   bm, err := Signatures([]Signature{Signature{NewFrame(BOF, Sequence{'p','d','f'}, 0, 0)}})
func Signatures(sigs []Signature, opts ...int) (*ByteMatcher, error) {
	b := newByteMatcher()
	b.Sigs = make([][]keyFrame, len(sigs))
	se := make(sigErrors, 0)
	var distance, rng, choices, varlen = 8192, 2049, 64, 1
	// override defaults if distance, range or choices values are given
	switch len(opts) {
	case 1:
		distance = opts[0]
	case 2:
		distance, rng = opts[0], opts[1]
	case 3:
		distance, rng, choices = opts[0], opts[1], opts[2]
	case 4:
		distance, rng, choices, varlen = opts[0], opts[1], opts[2], opts[3]
	}
	// process each of the sigs, adding them to b.Sigs and the various seq/frame/testTree sets
	for i, sig := range sigs {
		err := b.process(sig, i, distance, rng, choices, varlen)
		if err != nil {
			se = append(se, err)
		}
	}
	if len(se) > 0 {
		return b, se
	}
	// set the maximum distances for this test tree so can properly size slices for matching
	for _, t := range b.TestSet {
		t.MaxLeftDistance = maxLength(t.Left)
		t.MaxRightDistance = maxLength(t.Right)
	}
	// create aho corasick search trees for the lists of sequences (BOF, EOF and variable)
	b.bAho = ac.NewFixed(b.BofSeqs.Set)
	b.eAho = ac.NewFixed(b.EofSeqs.Set)
	b.vAho = ac.New(b.VarSeqs.Set)
	return b, nil
}

// Start initialises the aho corasick search trees after a Bytematcher has been loaded.
func (b *ByteMatcher) Start() {
	b.bAho = ac.NewFixed(b.BofSeqs.Set)
	b.eAho = ac.NewFixed(b.EofSeqs.Set)
	b.vAho = ac.New(b.VarSeqs.Set)
}

// Identify matches a Bytematcher's signatures against the input siegreader.Buffer.
// Results are passed on the returned int channel. These ints are the indexes of the matching signatures.
// A limit channel is also returned. This channel can be used to instruct the Bytematcher to ignore all matches except for the supplied slice of signature indexes.
//
// Example:
//   ret, limit := bm.Identify(buf)
//   for v := range ret {
//     if v == 0 {
//       limit <- []int{2,3}
//     }
//   }
func (b *ByteMatcher) Identify(sb *siegreader.Buffer) (chan int, chan []int) {
	ret, limit := make(chan int), make(chan []int, 50)
	go b.identify(sb, ret, limit)
	return ret, limit
}

// Stats returns information about the sending Bytematcher including the number of BOF, VAR and EOF sequences, the number of BOF and EOF frames, and the total number of tests.
func (b *ByteMatcher) Stats() string {
	str := fmt.Sprintf("BOF seqs: %v\n", len(b.BofSeqs.Set))
	str += fmt.Sprintf("EOF seqs: %v\n", len(b.EofSeqs.Set))
	str += fmt.Sprintf("BOF frames: %v\n", len(b.BofFrames.Set))
	str += fmt.Sprintf("EOF frames: %v\n", len(b.EofFrames.Set))
	str += fmt.Sprintf("VAR seqs: %v\n", len(b.VarSeqs.Set))
	str += fmt.Sprintf("Total Tests: %v\n", len(b.TestSet))
	var c, ic, l, r, ml, mr int
	for _, t := range b.TestSet {
		c += len(t.Complete)
		ic += len(t.Incomplete)
		l += len(t.Left)
		if ml < t.MaxLeftDistance {
			ml = t.MaxLeftDistance
		}
		r += len(t.Right)
		if mr < t.MaxRightDistance {
			mr = t.MaxRightDistance
		}
	}
	str += fmt.Sprintf("Complete Tests: %v\n", c)
	str += fmt.Sprintf("Incomplete Tests: %v\n", ic)
	str += fmt.Sprintf("Left Tests: %v\n", l)
	str += fmt.Sprintf("Right Tests: %v\n", r)
	str += fmt.Sprintf("Maximum Left Distance: %v\n", ml)
	str += fmt.Sprintf("Maximum Right Distance: %v\n", mr)
	return str
}

func newByteMatcher() *ByteMatcher {
	return &ByteMatcher{
		nil,
		make([]*testTree, 0),
		newSeqSet(),
		newSeqSet(),
		newSeqSet(),
		newFrameSet(),
		newFrameSet(),
		&ac.Ac{},
		&ac.Ac{},
		&ac.Ac{},
	}
}

type sigErrors []error

func (se sigErrors) Error() string {
	str := "bytematcher.Signatures errors:"
	for _, v := range se {
		str += v.Error()
		str += "\n"
	}
	return str
}

// Signature processing functions

func (b *ByteMatcher) processSeg(seg Signature, idx, i, l, x, y int, rev bool, ss *seqSet, fs *frameSet, fb, fe int) keyFrame {
	var k keyFrame
	var left, right []Frame
	if l > 0 {
		sequences := newSequencer(rev)
		var seqs [][]byte
		if rev {
			for j := y - 1; j >= x; j-- {
				seqs = sequences(seg[j])
			}

		} else {
			for _, f := range seg[x:y] {
				seqs = sequences(f)
			}
		}
		k, left, right = keyframe(seg, x, y)
		for _, seq := range seqs {
			hi := ss.add(seq, len(b.TestSet))
			if hi == len(b.TestSet) {
				b.TestSet = append(b.TestSet, newTestTree())
			}
			b.TestSet[hi].add([2]int{idx, i}, left, right)
		}
	} else {
		k, left, right = keyframe(seg, fb, fe)
		hi := fs.add(seg[fb], len(b.TestSet))
		if hi == len(b.TestSet) {
			b.TestSet = append(b.TestSet, newTestTree())
		}
		b.TestSet[hi].add([2]int{idx, i}, left, right)
	}
	return k
}

func (b *ByteMatcher) process(sig Signature, idx, distance, rng, choices, varlen int) error {
	// We start by segmenting the signature
	segs := segment(sig, distance, rng)
	// Our goal is to turn a signature into a set of keyframes
	kf := make([]keyFrame, len(segs))
	for i, seg := range segs {
		switch characterise(seg, distance) {
		case bofZero:
			// For a segment that is anchored at a 0 offset to the BOF, add its sequences to the BOF AC.
			l, x, y := bofLength(seg, choices)
			kf[i] = b.processSeg(seg, idx, i, l, x, y, false, b.BofSeqs, b.BofFrames, 0, 1)
		case eofZero:
			l, x, y := eofLength(seg, choices)
			kf[i] = b.processSeg(seg, idx, i, l, x, y, true, b.EofSeqs, b.EofFrames, len(seg)-1, len(seg))
		case bofWindow:
			l, x, y := varLength(seg, choices)
			kf[i] = b.processSeg(seg, idx, i, l, x, y, false, b.VarSeqs, b.BofFrames, 0, 1)
		case eofWindow:
			l, x, y := varLength(seg, choices)
			kf[i] = b.processSeg(seg, idx, i, l, x, y, false, b.VarSeqs, b.EofFrames, len(seg)-1, len(seg))
		default:
			l, x, y := varLength(seg, choices)
			if l < varlen {
				return fmt.Errorf("Variable offset segment encountered that can't be turned into a sequence: signature %d, segment %d", idx, i)
			}
			kf[i] = b.processSeg(seg, idx, i, l, x, y, false, b.VarSeqs, b.BofFrames, 0, 1)
		}
	}
	b.Sigs[idx] = kf
	return nil
}

func (m *matcher) sendStrike(s strike) bool {
	select {
	case <-m.quit:
		close(m.incoming)
		return false
	case m.incoming <- s:
		return true
	}
}

// Identify function

func (b *ByteMatcher) identify(buf *siegreader.Buffer, r chan int, l chan []int) {
	m := newMatcher(b, buf, r, l)
	go m.match()
	// Test any BOF Frames
	for i, f := range b.BofFrames.Set {
		if match, matches := f.Match(buf.MustSlice(0, TotalLength(f), false)); match {
			min, _ := f.Length()
			for _, off := range matches {
				if !m.sendStrike(strike{b.BofFrames.TestTreeIndex[i], off - min, min, false, true}) {
					return
				}
			}
		}
	}
	// Test any EOF Frames (at this stage, there shouldn't be any)
	for i, f := range b.EofFrames.Set {
		if match, matches := f.MatchR(buf.MustSlice(0, TotalLength(f), true)); match {
			for _, off := range matches {
				if !m.sendStrike(strike{b.EofFrames.TestTreeIndex[i], off, 0, true, true}) {
					return
				}
			}
		}
	}
	// Test absolute BOF sequences
	bchan := b.bAho.IndexFixed(buf.NewReader())
	for bi := range bchan {
		if !m.sendStrike(strike{b.BofSeqs.TestTreeIndex[bi], 0, len(b.BofSeqs.Set[bi]), false, false}) {
			return
		}
	}

	select {
	case <-m.quit:
		close(m.incoming)
		return
	default:
	}

	// Start the variable length aho-corasick search
	vchan := b.vAho.IndexQ(buf.NewReader(), m.quit)

	// Setup a reverse reader and start the EOF aho-corasick search
	rr, err := buf.NewReverseReader()
	if err != nil && err != io.EOF {
		panic(err)
	}
	echan := b.eAho.IndexFixedQ(rr, m.quit)

	for {
		select {
		case vi, ok := <-vchan:
			if !ok {
				vchan = nil
			} else {
				if !m.sendStrike(strike{b.VarSeqs.TestTreeIndex[vi.Index], vi.Offset, len(b.VarSeqs.Set[vi.Index]), false, false}) {
					return
				}
			}
		case ei, ok := <-echan:
			if !ok {
				echan = nil
			} else {
				if !m.sendStrike(strike{b.EofSeqs.TestTreeIndex[ei], 0, len(b.EofSeqs.Set[ei]), true, false}) {
					return
				}
			}
		}
		if vchan == nil && echan == nil {
			break
		}
	}
	close(m.incoming)
}
