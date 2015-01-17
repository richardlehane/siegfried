package siegreader

import "log"

type mmap struct {
	*file

	handle uintptr // used for windows unmap
	buf    []byte
}

func newMmap() interface{} {
	return &mmap{}
}

func (m *mmap) setSource(f *file) {
	m.file = f
	err := m.mapFile()
	if m.src.Name() == "testdata/skeleton-suite/fmt/fmt-355-signature-id-522.rtf" {
		log.Fatalf("mapped %d", len(m.buf))
	}
	if err != nil {
		log.Fatalf("Siegreader fatal error while attempting to mmap: %s; error: %v\n", m.src.Name(), err)
	}
}

func (m *mmap) slice(off int64, l int) []byte {
	/*if int(off)+l > len(m.buf) {
		log.Fatalf("illegal mmap access for %s, %d, %d, buf len %d", m.src.Name(), off, l, len(m.buf))
	}*/
	return m.buf[int(off) : int(off)+l]
}

func (m *mmap) eofSlice(off int64, l int) []byte {
	o := int(m.sz - off)
	return m.buf[o-l : o]
}

func (m *mmap) reset() {
	if m.buf == nil {
		return
	}
	err := m.unmap()
	if err != nil {
		log.Fatalf("Siegreader fatal error while unmapping: %s; error: %v\n", m.src.Name(), err)
	}
	m.buf = nil
	return
}
