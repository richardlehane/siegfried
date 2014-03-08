package siegreader

// Slice

func (b *Buffer) Slice(s, e int) ([]byte, error) {
	// block until the slice is available

}

// Reader

type Reader struct {
	i int
	b *Buffer
}

func (r *Buffer) NewReader() *Reader {
	// A BOF reader may not have been used, trigger a fill if necessary.
	return &Reader{0, b}
}

func (r *Reader) Read(b []byte) (int, error) {

}

func (r *Reader) ReadByte() (byte, error) {
	if b.w+readSz > len(b.buf) {
		b.cont <- struct{}{}
	}
	b
	r.i++
}

func (r *Reader) ReadAt(b []byte, off int64) (int, error) {
	// block until the read is possible
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	// block until the seek is possible

}

// BOF Reader

type BOFReader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewBOFReader() *BOFReader {
	// fill the BOF now, if not already done
	return &BOFReader{0, b}
}

func (r *BOFReader) Read(b []byte) (int, error) {

}

func (r *BOFReader) ReadByte() (byte, error) {

}

// Reverse Reader

type EOFReader struct {
	i int
	b *Buffer
}

func (b *Buffer) NewEOFReader() *ReverseReader {
	// fill the EOF now, if possible and not already done
	return &ReverseReader{0, b}
}

func (r *ReverseReader) Read(b []byte) (int, error) {

}

func (r *ReverseReader) ReadByte() (byte, error) {
	// block if a stream, not a file
	if r.eof != nil {
		<-r.eof
	}

	if r.i > readSz {
		return 0, io.EOF
	}
	r.i++
	if len(b.tail) > 0 {
		return b.tail[len(b.tail)-r.i], nil
	} else {
		return b.buf[len(b.buf)-r.i], nil
	}
}
