package bytematcher

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/richardlehane/ac"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Bytematcher struct {
	Sigs       [][]keyFrame
	Priorities [][]int // given a match sig X, should we await any other signatures with a greater priority?

	TestSet []*testTree

	BofSeqs *seqSet
	EofSeqs *seqSet
	VarSeqs *seqSet

	BofFrames *frameSet
	EofFrames *frameSet

	bAho *ac.Ac
	eAho *ac.Ac
	vAho *ac.Ac

	rdr *siegreader.Reader
}

func Signatures(sigs []Signature, opts ...int) (*Bytematcher, error) {
	b := newBytematcher()
	n := make([][]keyFrame, len(b.Sigs)+len(sigs))
	copy(n, b.Sigs)
	b.Sigs = n
	se := make(sigErrors, 0)
	var distance, rng, choices = 8192, 2048, 64
	switch len(opts) {
	case 1:
		distance = opts[0]
	case 2:
		distance, rng = opts[0], opts[1]
	case 3:
		distance, rng, choices = opts[0], opts[1], opts[2]
	}
	for i, sig := range sigs {
		err := b.process(sig, distance, rng, choices, i)
		if err != nil {
			se = append(se, err.(sigError))
		}
	}
	if len(se) > 0 {
		return b, se
	}
	b.bAho = ac.NewFixed(b.BofSeqs.Set)
	b.eAho = ac.NewFixed(b.EofSeqs.Set)
	b.vAho = ac.New(b.VarSeqs.Set)
	return b, nil
}

func (b *Bytematcher) SetPriorities(p [][]int) {
	b.Priorities = p
}

func (b *Bytematcher) Start() {
	b.bAho = ac.NewFixed(b.BofSeqs.Set)
	b.eAho = ac.NewFixed(b.EofSeqs.Set)
	b.vAho = ac.New(b.VarSeqs.Set)
}

func (b *Bytematcher) Identify(r *siegreader.Reader) chan int {
	ret := make(chan int)
	b.rdr = r
	go b.identify(ret)
	return ret
}

func (b *Bytematcher) Stats() string {
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
		if ml < maxLength(t.Left) {
			ml = maxLength(t.Left)
		}
		r += len(t.Right)
		if mr < maxLength(t.Right) {
			mr = maxLength(t.Right)
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

func (b *Bytematcher) KeyFrames() []string {
	strs := make([]string, len(b.Sigs))
	for i, sig := range b.Sigs {
		str := "\n["
		for _, kf := range sig[:len(sig)-1] {
			str += kf.String() + ", "
		}
		str += sig[len(sig)-1].String()
		str += "]"
		strs[i] = str
	}
	return strs
}

func newBytematcher() *Bytematcher {
	return &Bytematcher{
		make([][]keyFrame, 0),
		make([]*testTree, 0),
		newSeqSet(),
		newSeqSet(),
		newSeqSet(),
		newFrameSet(),
		newFrameSet(),
		&ac.Ac{},
		&ac.Ac{},
		&ac.Ac{},
		new(bytes.Buffer),
	}
}

type sigError int

func (se sigError) Error() string {
	return fmt.Sprintf("Problem with signature %v\n", se)
}

type sigErrors []sigError

func (se sigErrors) Error() string {
	str := "bytematcher.Signatures errors:"
	for _, i := range se {
		str += fmt.Sprintf("Problem with signature %v\n", i)
	}
	return str
}

func (b *Bytematcher) processSeg(seg Signature, idx, i, l, x, y int, rev bool, ss *seqSet, fs *frameSet, fb, fe int) keyFrame {
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

func (b *Bytematcher) process(sig Signature, distance, rng, choices, idx int) error {
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
			if l < 1 {
				return fmt.Errorf("Variable offset segment encountered that can't be turned into a sequence: signature %d, segment %d", idx, i)
			}
			kf[i] = b.processSeg(seg, idx, i, l, x, y, false, b.VarSeqs, b.BofFrames, 0, 1)
		}
	}
	b.Sigs[idx] = kf
	return nil
}
