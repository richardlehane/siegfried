package bytematcher

import . "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"

// Sequence Sets and Frame Sets

// As far as possible, signatures are flattened into simple byte sequences grouped into three sets: BOF, EOF and variable offset sets.
// When a byte sequence is matched, the TestTree is examined for keyframe matches and to conduct further tests.
type seqSet struct {
	Set           [][]byte
	TestTreeIndex []int          // the Set and TestTreeIndex slices are of equal length. The index of the []byte is used to look up the index of the relevant TestTree
	exists        map[string]int // this map is used for constructing the set but isn't needed after, hence not exported and not in the saved GOB file
}

func newSeqSet() *seqSet {
	return &seqSet{make([][]byte, 0), make([]int, 0), make(map[string]int)}
}

// Add sequence to set. Provides latest TestTreeIndex, returns actual TestTreeIndex for hit insertion.
func (ss *seqSet) add(seq []byte, hi int) int {
	i, ok := ss.exists[string(seq)]
	if ok {
		return ss.TestTreeIndex[i]
	}
	ss.Set = append(ss.Set, seq)
	ss.TestTreeIndex = append(ss.TestTreeIndex, hi)
	ss.exists[string(seq)] = len(ss.Set) - 1
	return hi
}

// Some signatures cannot be represented by simple byte sequences. The first or last frames from these sequences are added to the BOF or EOF frame sets.
// Like sequences, frame matches are referred to the TestTree for further testing.
type frameSet struct {
	Set           []Frame
	TestTreeIndex []int
}

func newFrameSet() *frameSet {
	return &frameSet{make([]Frame, 0), make([]int, 0)}
}

// Add frame to set. Provides current testerIndex, returns actual testerIndex for hit insertion.
func (fs *frameSet) add(f Frame, hi int) int {
	for i, f1 := range fs.Set {
		if f1.Equals(f) {
			return fs.TestTreeIndex[i]
		}
	}
	fs.Set = append(fs.Set, f)
	fs.TestTreeIndex = append(fs.TestTreeIndex, hi)
	return hi
}
