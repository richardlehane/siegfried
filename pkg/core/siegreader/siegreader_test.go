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

var testBytes = []byte(testString)

var testfile = filepath.Join("..", "..", "..", "cmd", "sf", "testdata", "benchmark", "Benchmark.docx")

var bufs = New()

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
