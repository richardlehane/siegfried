package process

import (
	"fmt"

	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

type Options struct {
	Distance  int
	Range     int
	Choices   int
	VarLength int
}

var Defaults = Options{
	Distance:  8192,
	Range:     2049,
	Choices:   64,
	VarLength: 1,
}

type Process struct {
	KeyFrames [][]keyFrame
	Tests     []*testTree
	BOFFrames *frameSet
	EOFFrames *frameSet
	BOFSeq    *seqSet
	EOFSeq    *seqSet
	MaxEOF    int
	Options
}

func New() *Process {
	return &Process{
		make([][]keyFrame, 0),
		make([]*testTree, 0),
		newFrameSet(),
		newFrameSet(),
		newSeqSet(),
		newSeqSet(),
		0,
		Defaults,
	}
}

func (p *Process) SetOptions(opts ...int) {
	switch len(opts) {
	case 1:
		p.Options.Distance = opts[0]
	case 2:
		p.Options.Distance, p.Options.Range = opts[0], opts[1]
	case 3:
		p.Options.Distance, p.Options.Range, p.Options.Choices = opts[0], opts[1], opts[2]
	case 4:
		p.Options.Distance, p.Options.Range, p.Options.Choices, p.Options.VarLength = opts[0], opts[1], opts[2], opts[3]
	}
}

func (p *Process) AddSignature(sig frames.Signature) error {
	segments := p.splitSegments(sig)
	kf := make([]keyFrame, len(segments))
	clstr := newCluster(p)
	for i, segment := range segments {
		var pos position
		c := characterise(segment)
		switch c {
		case bofZero:
			pos = bofLength(segment, p.Choices)
		case eofZero:
			pos = eofLength(segment, p.Choices)
		default:
			pos = varLength(segment, p.Choices)
		}
		if pos.length < 1 {
			switch c {
			case bofZero, bofWindow:
				kf[i] = p.addToFrameSet(segment, i, p.BOFFrames, 0, 1)
			case eofZero, eofWindow:
				kf[i] = p.addToFrameSet(segment, i, p.EOFFrames, len(segment)-1, len(segment))
			default:
				return fmt.Errorf("Variable offset segment encountered that can't be turned into a sequence: signature %d, segment %d", len(p.KeyFrames), i)
			}
		} else {
			switch c {
			case bofZero, bofWindow, bofWild:
				clstr = clstr.commit()
				kf[i] = clstr.add(segment, i, pos)
			case prev:
				kf[i] = clstr.add(segment, i, pos)
			case succ:
				if !clstr.rev {
					clstr = clstr.commit()
					clstr.rev = true
				}
				kf[i] = clstr.add(segment, i, pos)
			case eofZero, eofWindow, eofWild:
				if !clstr.rev {
					clstr = clstr.commit()
					clstr.rev = true
				}
				kf[i] = clstr.add(segment, i, pos)
				clstr = clstr.commit()
				clstr.rev = true
			}
		}
	}
	clstr.commit()
	updatePositions(kf)
	p.MaxEOF = maxEOF(p.MaxEOF, kf)
	p.KeyFrames = append(p.KeyFrames, kf)
	return nil
}

type cluster struct {
	rev    bool
	kfs    []keyFrame
	p      *Process
	w      wac.Seq
	ks     []int
	lefts  [][]frames.Frame
	rights [][]frames.Frame
}

func newCluster(p *Process) *cluster {
	c := new(cluster)
	c.kfs = make([]keyFrame, 0)
	c.p = p
	c.w.Choices = make([]wac.Choice, 0)
	c.ks = make([]int, 0)
	c.lefts = make([][]frames.Frame, 0)
	c.rights = make([][]frames.Frame, 0)
	return c
}

func (c *cluster) add(seg frames.Signature, i int, pos position) keyFrame {
	sequences := frames.NewSequencer(c.rev)
	k, left, right := toKeyFrame(seg, pos)
	c.kfs = append(c.kfs, k)
	var seqs [][]byte
	// do it all backwards
	if c.rev {
		for j := pos.end - 1; j >= pos.start; j-- {
			seqs = sequences(seg[j])
		}
		c.w.Choices = append([]wac.Choice{wac.Choice(seqs)}, c.w.Choices...)
		c.ks = append([]int{i}, c.ks...)
		c.lefts = append([][]frames.Frame{left}, c.lefts...)
		c.rights = append([][]frames.Frame{right}, c.rights...)
	} else {
		for _, f := range seg[pos.start:pos.end] {
			seqs = sequences(f)
		}
		c.w.Choices = append(c.w.Choices, wac.Choice(seqs))
		c.ks = append(c.ks, i)
		c.lefts = append(c.lefts, left)
		c.rights = append(c.rights, right)
	}
	return k
}

func (c *cluster) commit() *cluster {
	// commit nothing if the cluster is empty
	if len(c.w.Choices) == 0 {
		return newCluster(c.p)
	}
	updatePositions(c.kfs)
	c.w.MaxOffsets = make([]int, len(c.kfs))
	if c.rev {
		for i := range c.w.MaxOffsets {
			c.w.MaxOffsets[i] = c.kfs[len(c.kfs)-1-i].Key.PMax
		}
	} else {
		for i, v := range c.kfs {
			c.w.MaxOffsets[i] = v.Key.PMax
		}
	}
	var ss *seqSet
	if c.rev {
		ss = c.p.EOFSeq
	} else {
		ss = c.p.BOFSeq
	}
	hi := ss.add(c.w, len(c.p.Tests))
	l := len(c.ks)
	if hi == len(c.p.Tests) {
		for i := 0; i < l; i++ {
			c.p.Tests = append(c.p.Tests, newTestTree())
		}
	}
	for i := 0; i < l; i++ {
		c.p.Tests[hi+i].add([2]int{len(c.p.KeyFrames), c.ks[i]}, c.lefts[i], c.rights[i])
	}
	return newCluster(c.p)
}

func (p *Process) addToFrameSet(segment frames.Signature, i int, fs *frameSet, start, end int) keyFrame {
	k, left, right := toKeyFrame(segment, position{0, start, end})
	hi := fs.add(segment[start], len(p.Tests))
	if hi == len(p.Tests) {
		p.Tests = append(p.Tests, newTestTree())
	}
	p.Tests[hi].add([2]int{len(p.KeyFrames), i}, left, right)
	return k
}
