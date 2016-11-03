package siegreader

import "log"

type smallfile struct {
	*file

	buf [smallFileSz]byte
}

func newSmallFile() interface{} {
	return &smallfile{}
}

func (sf *smallfile) setSource(f *file) {
	sf.file = f
	i, err := sf.src.ReadAt(sf.buf[:], 0)
	if i != int(sf.sz) {
		log.Fatalf("Siegreader fatal error: failed to read %s, got %d bytes of %d, error: %v\n", sf.src.Name(), i, sf.sz, err)
	}
}

func (sf *smallfile) slice(off int64, l int) []byte {
	return sf.buf[int(off) : int(off)+l]
}

func (sf *smallfile) eofSlice(off int64, l int) []byte {
	o := int(sf.sz - off)
	return sf.buf[o-l : o]
}
