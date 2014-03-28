package siegreader

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var teststring = "abracadabra"

var testfile = filepath.Join("..", "..", "..", "cmd", "sieg", "testdata", "benchmark", "Benchmark.docx")

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Error("Buffer is nil")
	}
}

func setup(r io.Reader, t *testing.T) *Buffer {
	b := New()
	err := b.SetSource(r)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	return b
}

func TestStrSource(t *testing.T) {
	r := strings.NewReader(teststring)
	setup(r, t)
}

func TestFileSource(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	setup(r, t)
	r.Close()
}
