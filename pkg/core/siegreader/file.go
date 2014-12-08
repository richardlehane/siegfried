package siegreader

import "os"

// b) File
//    b i)   Satisifed with small read beginning and end
//    b ii)  Small enough for full read
//    b iii) Mmap
//    b iv) Too big for MMap - small buffers and expose ReaderAt
type file struct {
	bof  []byte
	date []byte
}

type mmapData struct {
	f *os.File
	b []byte
}
