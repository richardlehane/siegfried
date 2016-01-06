package siegreader

type source interface {
	IsSlicer() bool
	Slice(off int64, l int) ([]byte, error)
	EofSlice(off int64, l int) ([]byte, error)
	Size() int64
}

// an external buffer is a non-file stream that implements the Slice() etc. methods
// this is used to prevent unnecessary copying of webarchive WARC/ARC readers
type external struct{ source }

func newExternal() interface{} { return &external{} }

func (e *external) setSource(src source) error {
	e.source = src
	if e.Size() == 0 {
		return ErrEmpty
	}
	return nil
}

// SizeNow is a non-blocking Size().
func (e *external) SizeNow() int64 { return e.Size() }

func (e *external) CanSeek(off int64, whence bool) (bool, error) {
	if e.Size() < off {
		return false, nil
	}
	return true, nil
}
