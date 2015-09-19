package siegreader

type source interface {
	Slice(off int64, l int) ([]byte, error)
	EofSlice(off int64, l int) ([]byte, error)
	Size() int64
}

type external struct {
	quit    chan struct{}
	limited bool
	limit   chan struct{}
	source
}

func newExternal() interface{} { return &external{} }

func (e *external) setSource(src source) {
	e.quit = nil
	e.limited = false
	e.limit = nil
	e.source = src
}

func (e *external) SizeNow() int64 { return e.Size() }

func (e *external) Stream() bool { return false }

func (e *external) SetQuit(q chan struct{}) { e.quit = q }

func (e *external) setLimit() {
	e.limited = true
	e.limit = make(chan struct{})
}

func (e *external) waitLimit() {
	if e.limited {
		select {
		case <-e.limit:
		case <-e.quit:
		}
	}
}

func (e *external) hasQuit() bool {
	select {
	case <-e.quit:
		return true
	default:
	}
	return false
}

func (e *external) reachedLimit() { close(e.limit) }

func (e *external) canSeek(off int64, whence bool) (bool, error) {
	if e.Size() < off {
		return false, nil
	}
	return true, nil
}
