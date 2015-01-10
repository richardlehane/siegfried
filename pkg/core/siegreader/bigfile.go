package siegreader

import "sync"

// big file data
const (
	wheelSz = readSz * 16
	eofSz   = readSz * 2
)

type enc uint8

const (
	notEnc enc = iota
	wheelEnc
	eofEnc
)

// A big file gets read into a sliding window
// If slices are requested, they are copied into new slices (either from the window,  or by exposing underlying reader at)
// should jump into the EOF slice as well (to cover PDF/A etc which can have markers near end)
// If an adjacent slice of size readSz is requested, it is copied into the wheel. Otherwise it is made fresh.
// Possibility of collision (two clients requesting adjacent windows, causing race) unlikely because of size of wheel (i.e. may get one or two collisions but unlikely to get 16 in a row)
type bigfile struct {
	*file

	mu   sync.Mutex
	i    int   // wheel offset for next write
	last int64 // file offset of last write to allow test for adjacency

	broken bool // is the wheel broken (two legitimate adjacent requests)?
	eof    [eofSz]byte
	wheel  [wheelSz]byte
}

func (b *bigfile) adjacent(o int64, l int) bool {
	return l == readSz && o == b.last+int64(readSz)
}

// is the requested slice in the wheel or the eof?
func (b *bigfile) enclosed(o int64, l int) enc {
	if int(b.sz-o) <= eofSz {
		return eofEnc
	}
	if o < b.last && l <= wheelSz && int(b.last-o) <= l {
		return wheelEnc
	}
	return notEnc
}

func (b *bigfile) slice(o int64, l int) ([]byte, error) {
	if b.adjacent(o, l) {
		// safe to send direct reference
		return b.wheel[:], nil
	}
	ret := make([]byte, l)
	// if within the wheel copy
	switch b.enclosed(o, l) {
	case eofEnc:
	case wheelEnc:
		copy(ret, b.wheel[:])
		return ret, nil
	}
	// otherwise we just expose the underlying reader at
	b.src.ReadAt(ret, o)
	return ret, nil
}
