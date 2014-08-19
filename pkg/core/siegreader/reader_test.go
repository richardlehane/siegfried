package siegreader

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	b := setup(strings.NewReader(teststring), t)
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

func readAt(t *testing.T, r *Reader) {
	buf := make([]byte, 5)
	i, err := r.ReadAt(buf, 4)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if i != 5 {
		t.Errorf("Read error: expecting a read length of 5, got %v", i)
	}
	if string(buf) != "cadab" {
		t.Errorf("Read error: %v should equal %v", buf, "cadab")
	}
}

func TestReadAt(t *testing.T) {
	b := setup(strings.NewReader(teststring), t)
	r := b.NewReader()
	readAt(t, r)
}

func readByte(t *testing.T, r *Reader) {
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

func TestReadByte(t *testing.T) {
	b := setup(strings.NewReader(teststring), t)
	r := b.NewReader()
	readByte(t, r)
}

func seek(t *testing.T, r *Reader) {
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

func TestSeek(t *testing.T) {
	b := setup(strings.NewReader(teststring), t)
	r := b.NewReader()
	seek(t, r)
}

func TestReuse(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	r.Close()
	nr := strings.NewReader(teststring)
	q := make(chan struct{})
	err = b.SetSource(nr)
	b.SetQuit(q)
	if err != nil && err != io.EOF {
		t.Errorf("Read error: %v", err)
	}
	reuse := b.NewReader()
	readByte(t, reuse)
	seek(t, reuse)
}

func drain(r io.ByteReader, results chan int) {
	var i int
	for _, e := r.ReadByte(); e == nil; _, e = r.ReadByte() {
		i++
	}
	results <- i
}

func TestDrain(t *testing.T) {
	b := setup(strings.NewReader(teststring), t)
	r := b.NewReader()
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 11 {
		t.Errorf("Expecting 11, got %v", i)
	}
}

func TestDrainFile(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	first := b.NewReader()
	results := make(chan int)
	go drain(first, results)
	if i := <-results; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	r.Close()
}

func TestMultiple(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	first := b.NewReader()
	second := b.NewReader()
	results := make(chan int)
	go drain(first, results)
	go drain(second, results)
	if i := <-results; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	if i := <-results; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
}

func TestReverse(t *testing.T) {
	b := setup(strings.NewReader(teststring), t)
	r, err := b.NewReverseReader()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	first := b.NewReader()
	results := make(chan int)
	go drain(first, results)
	<-results
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
	if c != 'r' {
		t.Errorf("Read error: expecting 'r', got %v", c)
	}
}

func TestReverseDrainFile(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	quit := make(chan struct{})
	b.SetQuit(quit)
	first := b.NewReader()
	firstResults := make(chan int, 24040)
	last, _ := b.NewReverseReader()
	lastResults := make(chan int)
	go drain(first, firstResults)
	go drain(last, lastResults)
	if i := <-lastResults; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	r.Close()
}
