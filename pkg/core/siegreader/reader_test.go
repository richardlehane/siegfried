package siegreader

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := b.NewReader()
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
	r := b.NewReader()
	readAt(t, r)
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
	if c != '6' {
		t.Errorf("Read error: expecting '6', got %s", string(c))
	}

}

func TestSeek(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
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
	nr := strings.NewReader(testString)
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
	b := setup(strings.NewReader(testString), t)
	r := b.NewReader()
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 62 {
		t.Errorf("Expecting 62, got %v", i)
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
	b := setup(strings.NewReader(testString), t)
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
	if c != 'Z' {
		t.Errorf("Read error: expecting 'Z', got %s", string(c))
	}
	c, err = r.ReadByte()
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	if c != 'Y' {
		t.Errorf("Read error: expecting 'Y', got %s", string(c))
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

func TestLimit(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r := b.NewLimitReader(5)
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 5 {
		t.Errorf("Expecting 5, got %d", i)
	}
}

func TestReverseLimit(t *testing.T) {
	b := setup(strings.NewReader(testString), t)
	r, err := b.NewLimitReverseReader(5)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}
	results := make(chan int)
	go drain(r, results)
	if i := <-results; i != 5 {
		t.Errorf("Expecting 5, got %d", i)
	}
}
