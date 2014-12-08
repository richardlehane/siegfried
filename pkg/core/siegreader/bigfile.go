package siegreader

// b) File
//    b i)   Satisifed with small read beginning and end
//    b ii)  Small enough for full read
//    b iii) Mmap
//    b iv) Too big for MMap - small buffers and expose ReaderAt
type bigfile struct {
	bof []byte // generous 256Kb
	eof []byte // generous 140Kb
}

func (b *bigfile) Slice(o, l int) ([]byte, error) {
	return nil, nil
}
