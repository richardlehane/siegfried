package siegreader

// big file data

const (
	bfWindow = 256000 // generous 256Kb
)

// A big file gets read into a sliding window
// If slices are requested, they are copied into new slices (either from the window,  or by exposing underlying reader at)
// should jump into the EOF slice as well (to cover PDF/A etc which can have markers near end)
type bigfile struct {
	window []byte
	eof    []byte
}

func (b *bigfile) Slice(o, l int) ([]byte, error) {
	return nil, nil
}
