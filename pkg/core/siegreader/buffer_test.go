package siegreader

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var teststring = "abracadabra"

var _ = filepath.Join("..", "..", "..", "cmd", "siegfried", "testdata", "benchmark", "Benchmark")

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Error("Buffer is nil")
	}
}

func setup(r io.Reader) *Buffer {
	b := New()
	err := b.SetSource(r)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	return b
}

func TestStrSource(t *testing.T) {

}

func TestFileSource(t *testing.T) {

}
