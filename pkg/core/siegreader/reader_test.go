package siegreader

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	b := setup(strings.NewReader(teststring))
	r := b.NewReader()
	buf := make([]byte, 11)
	i, err := r.Read(buf)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if i != 11 {
		t.Errorf("Read error: expecting a read length of 11, got %v", i)
	}
	if string(buf) != teststring {
		t.Errorf("Read error: %v should equal %v", buf, teststring)
	}
}

func TestReadAt(t *testing.T) {
	b := setup(strings.NewReader(teststring))
	r := b.NewReader()
	buf := make([]byte, 5)
	i, err := r.ReadAt(buf, 4)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if i != 5 {
		t.Errorf("Read error: expecting a read length of 5, got %v", i)
	}
	if string(buf) != "cadab" {
		t.Errorf("Read error: %v should equal %v", buf, teststring)
	}
}

func TestReadByte(t *testing.T) {
	b := setup(strings.NewReader(teststring))
	r := b.NewReader()
	c, err := r.ReadByte()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if c != 'a' {
		t.Errorf("Read error: expecting 'a', got %v", c)
	}
	c, err = r.ReadByte()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if c != 'b' {
		t.Errorf("Read error: expecting 'b', got %v", c)
	}
}

func TestSeek(t *testing.T) {
	b := setup(strings.NewReader(teststring))
	r := b.NewReader()
	_, err := r.Seek(6, 0)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	c, err := r.ReadByte()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if c != 'd' {
		t.Errorf("Read error: expecting 'd', got %v", c)
	}
}
