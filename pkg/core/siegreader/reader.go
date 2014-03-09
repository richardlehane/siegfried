package siegreader

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
	slc, err := r.b.Slice(r.i, len(b))
	copy(b, slc)
	return len(slc), err
}

func (r *Reader) ReadByte() (byte, error) {

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

func (b *Buffer) NewEOFReader() *EofReader {
	// fill the EOF now, if possible and not already done
	return &EofReader{0, b}
}

func (r *ReverseReader) Read(b []byte) (int, error) {

}

func (r *ReverseReader) ReadByte() (byte, error) {

}
