package bytematcher

import "io"

type reverseReader struct {
	i int
	b []byte
}

func (r *reverseReader) ReadByte() (byte, error) {
	if r.i < 1 {
		return 0, io.EOF
	}
	r.i--
	return r.b[r.i], nil
}

func newReverseReader(buf []byte) *reverseReader {
	return &reverseReader{len(buf), buf}
}
