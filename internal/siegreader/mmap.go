package siegreader

import "log"

type mmap struct {
	*file

	handle uintptr // for windows unmap
	buf    []byte
}

func newMmap() interface{} {
	return &mmap{}
}

func (m *mmap) setSource(f *file) error {
	m.file = f
	return m.mapFile()
}

func (m *mmap) slice(off int64, l int) []byte {
	return m.buf[int(off) : int(off)+l]
}

func (m *mmap) eofSlice(off int64, l int) []byte {
	o := int(m.sz - off)
	return m.buf[o-l : o]
}

func (m *mmap) reset() {
	err := m.unmap()
	if err != nil {
		log.Fatalf("Siegfried: fatal error while unmapping: %s; error: %v\n", m.src.Name(), err) // not polite of this package to panic - consider deprecate
	}
	m.buf = nil
	return
}
