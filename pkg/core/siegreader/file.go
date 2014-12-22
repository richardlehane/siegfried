package siegreader

import "os"

// b) File
//    b i)   Satisifed with small read beginning
//    b ii)  Small enough for full read
//    b iii) Mmap
//    b iv) Too big for MMap - small buffers and expose ReaderAt
type data interface {
}

type file struct {
	bof  []byte
	data data
}

type mmapData struct {
	f *os.File
	b []byte
}
