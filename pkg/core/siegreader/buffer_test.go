package siegreader

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var teststring = "abracadabra"

var testbytes = []byte("test12345678910YNESSjunktestyjunktestytest12345678910111223")

var testfile = filepath.Join("..", "..", "..", "cmd", "sieg", "testdata", "benchmark", "Benchmark.docx")

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Error("Buffer is nil")
	}
}

func setup(r io.Reader, t *testing.T) *Buffer {
	b := New()
	q := make(chan struct{})
	b.SetQuit(q)
	err := b.SetSource(r)
	if err != nil && err != io.EOF {
		t.Errorf("Read error: %v", err)
	}
	return b
}

func TestStrSource(t *testing.T) {
	r := strings.NewReader(teststring)
	b := setup(r, t)
	if b.Size() != len(teststring) {
		t.Error("String read: size error")
	}
}

func TestBytSource(t *testing.T) {
	r := bytes.NewBuffer(testbytes)
	b := setup(r, t)
	if b.Size() != len(testbytes) {
		t.Error("String read: size error")
	}
}

func TestFileSource(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	setup(r, t)
	r.Close()
}
