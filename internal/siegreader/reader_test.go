package siegreader

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := ReaderFrom(b)
	buf := make([]byte, 62)
	i, err := r.Read(buf)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if i != 62 {
		t.Errorf("Read error: expecting a read length of 62, got %v", i)
	}
	if string(buf) != testString {
		t.Errorf("Read error: %s should equal %s", string(buf), testString)
	}
	bufs.Put(b)
}

func readAt(t *testing.T, r *Reader) {
	buf := make([]byte, 5)
	i, err := r.ReadAt(buf, 4)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if i != 5 {
		t.Errorf("Read error: expecting a read length of 5, got %d", i)
	}
	if string(buf) != "45678" {
		t.Errorf("Read error: %s should equal %s", string(buf), "45678")
	}
}

func TestReadAt(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := ReaderFrom(b)
	readAt(t, r)
	bufs.Put(b)
}

func readByte(t *testing.T, r *Reader) {
	c, err := r.ReadByte()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if c != '0' {
		t.Errorf("Read error: expecting '0', got %s", string(c))
	}
	c, err = r.ReadByte()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if c != '1' {
		t.Errorf("Read error: expecting '1', got %s", string(c))
	}
}

func TestReadByte(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := ReaderFrom(b)
	readByte(t, r)
	bufs.Put(b)
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
	if c != '6' {
		t.Errorf("Read error: expecting '6', got %s", string(c))
	}

}

func TestSeek(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := ReaderFrom(b)
	seek(t, r)
	bufs.Put(b)
}

func TestReuse(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b, err := bufs.Get(r)
	if err != nil {
		t.Fatal(err)
	}
	bufs.Put(b)
	r.Close()
	nr := strings.NewReader(testString)
	q := make(chan struct{})
	b, err = bufs.Get(nr)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	b.Quit = q
	if err != nil && err != io.EOF {
		t.Errorf("Read error: %v", err)
	}
	reuse := ReaderFrom(b)
	readByte(t, reuse)
	seek(t, reuse)
	bufs.Put(b)
}

func drain(r io.ByteReader, results chan int) {
	var i int
	for _, e := r.ReadByte(); e == nil; _, e = r.ReadByte() {
		i++
	}
	results <- i
}

func TestDrain(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := ReaderFrom(b)
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 62 {
		t.Errorf("Expecting 62, got %v", i)
	}
	bufs.Put(b)
}

func TestDrainFile(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	first := ReaderFrom(b)
	results := make(chan int)
	go drain(first, results)
	if i := <-results; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	r.Close()
	bufs.Put(b)
}

func TestMultiple(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	first := ReaderFrom(b)
	second := ReaderFrom(b)
	results := make(chan int)
	go drain(first, results)
	go drain(second, results)
	if i := <-results; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	if i := <-results; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	bufs.Put(b)
}

func TestReverse(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := ReverseReaderFrom(b)
	first := ReaderFrom(b)
	results := make(chan int)
	go drain(first, results)
	<-results
	c, err := r.ReadByte()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if c != 'Z' {
		t.Fatalf("Read error: expecting 'Z', got %s", string(c))
	}
	c, err = r.ReadByte()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if c != 'Y' {
		t.Fatalf("Read error: expecting 'Y', got %s", string(c))
	}
	bufs.Put(b)
}

func TestReverseDrainFile(t *testing.T) {
	r, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	quit := make(chan struct{})
	b.Quit = quit
	first := ReaderFrom(b)
	firstResults := make(chan int, 1)
	last := ReverseReaderFrom(b)
	lastResults := make(chan int)
	go drain(first, firstResults)
	go drain(last, lastResults)
	if i := <-lastResults; i != 24040 {
		t.Errorf("Expecting 24040, got %v", i)
	}
	r.Close()
	bufs.Put(b)
}

func TestLimit(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := LimitReaderFrom(b, 5)
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 5 {
		t.Errorf("Expecting 5, got %d", i)
	}
	bufs.Put(b)
}

func TestReverseLimit(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	l := LimitReaderFrom(b, -1)
	firstResults := make(chan int, 1)
	go drain(l, firstResults)
	r := LimitReverseReaderFrom(b, 5)
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 5 {
		t.Errorf("Expecting 5, got %d", i)
	}
	bufs.Put(b)
}
