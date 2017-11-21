package siegreader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
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

func TestNewBufferPool(t *testing.T) {
	b := New()
	if b == nil {
		t.Error("Failed to make a new Buffer pool nil")
	}
}

func setup(r io.Reader, t *testing.T) *Buffer {
	buf, err := bufs.Get(r)
	if err != nil && err != io.EOF {
		t.Fatalf("Read error: %v", err)
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
	defer bufs.Put(b)
	b.Slice(0, readSz)
	if b.Size() != int64(len(testString)) {
		t.Errorf("String read: size error, expecting %d got %d", b.Size(), int64(len(testString)))
	}
}

func TestBytSource(t *testing.T) {
	r := bytes.NewBuffer(testBytes)
	b := setup(r, t)
	defer bufs.Put(b)
	b.Slice(0, readSz)
	if b.Size() != int64(len(testBytes)) {
		t.Error("String read: size error")
	}
	if len(b.Bytes()) != len(testBytes) {
		t.Error("String read: Bytes() error")
	}
}

func TestMMAPFile(t *testing.T) {
	r, err := os.Open(testfile)
	defer r.Close()
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	defer bufs.Put(b)
	stat, _ := r.Stat()
	if len(b.Bytes()) != int(stat.Size()) {
		t.Error("File read: Bytes() error")
	}
}

func TestBigFile(t *testing.T) {
	f, err := os.Open(testBigFile)
	defer f.Close()
	if err != nil {
		t.Fatal(err)
	}
	b := setup(f, t)
	defer bufs.Put(b)
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
}

func TestSmallFile(t *testing.T) {
	r, err := os.Open(testSmallFile)
	defer r.Close()
	if err != nil {
		t.Fatal(err)
	}
	b := setup(r, t)
	defer bufs.Put(b)
	b.setsmallfile()
	stat, _ := r.Stat()
	if len(b.Bytes()) != int(stat.Size()) {
		t.Error("File read: Bytes() error")
	}
}

// The following tests generate temp files filled with random data and compare io.ReadAt()
// calls with Slice() and EofSlice() calls for the various Buffer types (mmap file, big file,
// small file, big stream and small stream).

func TestMMAPFileRand(t *testing.T) {
	tf, err := makeTmp(100000)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()
	b := setup(tf, t)
	defer bufs.Put(b)
	if err := testBuffer(t, 1000, tf, b); err != nil {
		t.Fatal(err)
	}
}

func TestSmallFileRand(t *testing.T) {
	tf, err := makeTmp(10000)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()
	b := setup(tf, t)
	b.setsmallfile()
	defer bufs.Put(b)
	err = testBuffer(t, 10, tf, b)
	if err := testBuffer(t, 1000, tf, b); err != nil {
		t.Fatal(err)
	}
}

func TestBigFileRand(t *testing.T) {
	tf, err := makeTmp(100000)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()
	b := setup(tf, t)
	b.setbigfile()
	defer bufs.Put(b)
	if err := testBuffer(t, 1000, tf, b); err != nil {
		t.Fatal(err)
	}
}

func TestSmallStreamRand(t *testing.T) {
	var sz int64 = 100000
	tf, err := makeTmp(sz)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()
	lr := io.LimitReader(tf, sz)
	b := setup(lr, t)
	defer bufs.Put(b)
	if err := testBuffer(t, 1000, tf, b); err != nil {
		t.Fatal(err)
	}
}

func TestBigStreamRand(t *testing.T) {
	var sz int64 = 1000000000 // 1GB
	tf, err := makeTmp(sz)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()
	lr := io.LimitReader(tf, sz)
	b := setup(lr, t)
	defer bufs.Put(b)
	if err := testBuffer(t, 1000, tf, b); err != nil {
		t.Fatal(err)
	}
}

func makeTmp(sz int64) (f *os.File, err error) {
	tf, err := ioutil.TempFile("", "sftest")
	if err != nil {
		return nil, err
	}
	rnd := rand.New(rand.NewSource(rand.Int63()))
	wr, err := io.CopyN(tf, rnd, sz)
	if wr != sz {
		return nil, fmt.Errorf("didn't write rands successfully: tried %d, got %d", sz, wr)
	}
	if err != nil {
		return nil, err
	}
	nm := tf.Name()
	tf.Close() // force the flush
	return os.Open(nm)
}

func eofOff(sz, off int64, l int) (int64, int) {
	bof := sz - off - int64(l)
	if bof >= 0 {
		return bof, l
	}
	return 0, l + int(bof)
}

func joinErrs(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	strs := make([]string, len(errs))
	for i := range strs {
		strs[i] = errs[i].Error()
	}
	return fmt.Errorf("Got %d errors: %s", len(strs), strings.Join(strs, "\n"))
}

func testBuffer(t *testing.T, checks int, tf *os.File, bs *Buffer) error {
	stat, err := tf.Stat()
	if err != nil {
		return err
	}
	sz := stat.Size()
	lens := make([][]byte, 10)
	for i := range lens {
		lens[i] = make([]byte, rand.Intn(10000))
	}
	offs := make([]int64, checks)
	for i := range lens {
		offs[i] = rand.Int63n(sz)
	}
	var errs []error
	// test Slice()
	for _, o := range offs {
		rb := lens[rand.Intn(len(lens))]
		t.Logf("trying read and slice, at off %d and len %d\n", o, len(rb))
		wi, rerr := tf.ReadAt(rb, o)
		slc, serr := bs.Slice(o, len(rb))
		if !bytes.Equal(rb[:wi], slc) {
			var samplea, sampleb []byte
			if wi >= 3 {
				samplea = rb[:3]
			}
			if len(slc) >= 3 {
				sampleb = slc[:3]
			}
			errs = append(errs, fmt.Errorf("Bad Slice() read at offset %d, len %d; got %v & %v, errs reported %v & %v", o, len(rb), samplea, sampleb, rerr, serr))
		}
	}
	if bsz := bs.SizeNow(); bsz != sz {
		errs = append(errs, fmt.Errorf("SizeNow() does not match: got %d, expecting %d", bsz, sz))
	}
	// test EofSlice()
	for _, o := range offs {
		rb := lens[rand.Intn(len(lens))]
		off, l := eofOff(sz, o, len(rb))
		t.Logf("trying read and eofslice at EOF offset %d (real %d), len %d (real %d)\n", o, off, len(rb), l)
		wi, rerr := tf.ReadAt(rb[:l], off)
		slc, serr := bs.EofSlice(o, len(rb))
		if !bytes.Equal(rb[:wi], slc) {
			var samplea, sampleb []byte
			if wi >= 3 {
				samplea = rb[:3]
			}
			if len(slc) >= 3 {
				sampleb = slc[:3]
			}
			errs = append(errs, fmt.Errorf("Bad EofSlice() read at EOF offset %d (real %d), len %d (real %d); got %v & %v, errs reported %v & %v", o, off, len(rb), l, samplea, sampleb, rerr, serr))
		}
	}
	return joinErrs(errs)
}
