package siegreader

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testString = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	testBytes     = []byte(testString)
	testfile      = filepath.Join("..", "..", "cmd", "sf", "testdata", "benchmark", "Benchmark.docx")
	testBigFile   = filepath.Join("..", "..", "cmd", "sf", "testdata", "benchmark", "Benchmark.xml")
	testSmallFile = filepath.Join("..", "..", "cmd", "sf", "testdata", "benchmark", "Benchmark.gif")

	bufs = New()
)

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Error("New is nil")
	}
}

func setup(r io.Reader, t *testing.T) *Buffer {
	buf, err := bufs.Get(r)
	if err != nil && err != io.EOF {
		t.Errorf("Read error: %v", err)
	}
	q := make(chan struct{})
	buf.Quit = q
	return buf
}

func (b *Buffer) setbigfile() {
	b.bufferSrc.(*file).once.Do(func() {
		b.bufferSrc.(*file).data = b.bufferSrc.(*file).pool.bfpool.get().(*bigfile)
		b.bufferSrc.(*file).data.(*bigfile).setSource(b.bufferSrc.(*file))
	})
}

func (b *Buffer) setsmallfile() {
	b.bufferSrc.(*file).once.Do(func() {
		b.bufferSrc.(*file).data = b.bufferSrc.(*file).pool.sfpool.get().(*smallfile)
		b.bufferSrc.(*file).data.(*smallfile).setSource(b.bufferSrc.(*file))
	})
}

func TestStrSource(t *testing.T) {
	r := strings.NewReader(testString)
	b := setup(r, t)
	b.Slice(0, readSz)
	if b.Size() != int64(len(testString)) {
		t.Errorf("String read: size error, expecting %d got %d", b.Size(), int64(len(testString)))
	}
	bufs.Put(b)
}

func TestBytSource(t *testing.T) {
	r := bytes.NewBuffer(testBytes)
	b := setup(r, t)
	b.Slice(0, readSz)
	if b.Size() != int64(len(testBytes)) {
		t.Error("String read: size error")
	}
	if len(b.Bytes()) != len(testBytes) {
		t.Error("String read: Bytes() error")
	}
	bufs.Put(b)
}

func TestFileSource(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	stat, _ := r.Stat()
	if len(b.Bytes()) != int(stat.Size()) {
		t.Error("File read: Bytes() error")
	}
	r.Close()
	bufs.Put(b)
}

func TestBigSource(t *testing.T) {
	f, err := os.Open(testBigFile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(f, t)
	b.setbigfile()
	stat, _ := f.Stat()
	if len(b.Bytes()) != int(stat.Size()) {
		t.Error("File read: Bytes() error")
	}
	r := ReaderFrom(b)
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != int(stat.Size()) {
		t.Errorf("Expecting %d, got %d", int(stat.Size()), i)
	}
	f.Close()
	bufs.Put(b)
}

func TestSmallSource(t *testing.T) {
	r, err := os.Open(testSmallFile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	b.setsmallfile()
	stat, _ := r.Stat()
	if len(b.Bytes()) != int(stat.Size()) {
		t.Error("File read: Bytes() error")
	}
	r.Close()
	bufs.Put(b)
}
