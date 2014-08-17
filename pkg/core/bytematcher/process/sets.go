package process

import (
	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// Sequence Sets and Frame Sets

// As far as possible, signatures are flattened into simple byte sequences grouped into two sets: BOF and EOF sets.
// When a byte sequence is matched, the TestTree is examined for keyframe matches and to conduct further tests.
type seqSet struct {
	Set           []wac.Seq
	TestTreeIndex []int // The index of the testTree for the first choices. For subsequence choices, add the index of that choice to the test tree index.
}

func newSeqSet() *seqSet {
	return &seqSet{make([]wac.Seq, 0), make([]int, 0)}
}

// helper funcs to test equality of wac.Seq
func choiceExists(a []byte, b wac.Choice) bool {
	for _, v := range b {
		if string(a) == string(v) {
			return true
		}
	}
	return false
}

func seqEquals(a wac.Seq, b wac.Seq) bool {
	if a.Max != b.Max || len(a.Choices) != len(b.Choices) {
		return false
	}
	for i := range a.Choices {
		if len(a.Choices[i]) != len(b.Choices[i]) {
			return false
		}
		for _, v := range a.Choices[i] {
			if !choiceExists(v, b.Choices[i]) {
				return false
			}
		}
	}
	return true
}

func (ss *seqSet) exists(seq wac.Seq) (int, bool) {
	for i, v := range ss.Set {
		if seqEquals(seq, v) {
			return i, true
		}
	}
	return -1, false
}

// Add sequence to set. Provides latest TestTreeIndex, returns actual TestTreeIndex for hit insertion.
func (ss *seqSet) add(seq wac.Seq, hi int) int {
	i, ok := ss.exists(seq)
	if ok {
		return ss.TestTreeIndex[i]
	}
	ss.Set = append(ss.Set, seq)
	ss.TestTreeIndex = append(ss.TestTreeIndex, hi)
	return hi
}

// Some signatures cannot be represented by simple byte sequences. The first or last frames from these sequences are added to the BOF or EOF frame sets.
// Like sequences, frame matches are referred to the TestTree for further testing.
type frameSet struct {
	Set           []frames.Frame
	TestTreeIndex []int
}

func newFrameSet() *frameSet {
	return &frameSet{make([]frames.Frame, 0), make([]int, 0)}
}

// Add frame to set. Provides current testerIndex, returns actual testerIndex for hit insertion.
func (fs *frameSet) add(f frames.Frame, hi int) int {
	for i, f1 := range fs.Set {
		if f1.Equals(f) {
			return fs.TestTreeIndex[i]
		}
	}
	fs.Set = append(fs.Set, f)
	fs.TestTreeIndex = append(fs.TestTreeIndex, hi)
	return hi
}

type Fsmatch struct {
	Idx    int
	Off    int
	Length int
}

func (fs *frameSet) Index(buf *siegreader.Buffer, rev bool, quit chan struct{}) chan Fsmatch {
	ret := make(chan Fsmatch)
	go func() {
		var i int
		for {
			select {
			case <-quit:
				close(ret)
				return
			default:
			}
			if i >= len(fs.Set) {
				close(ret)
				return
			}
			f := fs.Set[i]
			var match bool
			var matches []int
			if rev {
				slc, err := buf.EofSlice(0, frames.TotalLength(f))
				if err != nil {
					close(ret)
					return
				}
				match, matches = f.MatchR(slc)
			} else {
				match, matches = f.Match(buf.MustSlice(0, frames.TotalLength(f), false))
			}
			if match {
				var min int
				if !rev {
					min, _ = f.Length()
				}
				for _, off := range matches {
					ret <- Fsmatch{i, off - min, min}
				}
			}
			i++
		}
		close(ret)
	}()
	return ret
}
