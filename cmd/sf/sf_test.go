package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	testhome = flag.String("testhome", "../roy/data", "override the default home directory")
	testdata = flag.String("testdata", filepath.Join(".", "testdata"), "override the default test data directory")
)

var s *siegfried.Siegfried

func setup(opts ...config.Option) error {
	if opts == nil && s != nil {
		return nil
	}
	var err error
	s = siegfried.New()
	config.SetHome(*testhome)
	opts = append(opts, config.SetDoubleUp())
	p, err := pronom.New(opts...)
	if err != nil {
		return err
	}
	return s.Add(p)
}

func identifyT(s *siegfried.Siegfried, p string) ([]string, error) {
	ids := make([]string, 0)
	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v, got: %v", p, err)
	}
	c, err := s.Identify(p, file)
	if c == nil {
		return nil, fmt.Errorf("failed to identify %v, got: %v", p, err)
	}
	for i := range c {
		ids = append(ids, i.String())
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func multiIdentifyT(s *siegfried.Siegfried, r string) ([][]string, error) {
	set := make([][]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *nr && path != r {
				return filepath.SkipDir
			}
			return nil
		}
		ids, err := identifyT(s, path)
		if err != nil {
			return err
		}
		set = append(set, ids)
		return nil
	}
	err := filepath.Walk(r, wf)
	return set, err
}

func check(i string, j []string) bool {
	for _, v := range j {
		if i == v {
			return true
		}
	}
	return false
}

func matchString(i []string) string {
	str := "[ "
	for _, v := range i {
		str += v
		str += " "
	}
	return str + "]"
}

func TestSuite(t *testing.T) {
	err := setup()
	if err != nil {
		t.Error(err)
	}
	expect := make([]string, 0)
	names := make([]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		last := strings.Split(path, string(os.PathSeparator))
		path = last[len(last)-1]
		var idx int
		idx = strings.Index(path, "container")
		if idx < 0 {
			idx = strings.Index(path, "signature")
		}
		if idx < 0 {
			idx = len(path)
		}
		strs := strings.Split(path[:idx-1], "-")
		if len(strs) == 2 {
			expect = append(expect, strings.Join(strs, "/"))
		} else if len(strs) == 3 {
			expect = append(expect, "x-fmt/"+strs[2])
		} else {
			return errors.New("long string encountered: " + path)
		}
		names = append(names, path)
		return nil
	}
	suite := filepath.Join(*testdata, "skeleton-suite")
	err = filepath.Walk(suite, wf)
	if err != nil {
		t.Fatal(err)
	}
	matches, err := multiIdentifyT(s, suite)
	if err != nil {
		t.Fatal(err)
	}
	if len(expect) != len(matches) {
		t.Error("Expect should equal matches")
	}
	var iter int
	for i, v := range expect {
		if !check(v, matches[i]) {
			t.Errorf("Failed to match signature %v; got %v; expected %v", names[i], matchString(matches[i]), v)

		} else {
			iter++
		}
	}
	if iter != len(expect) {
		t.Errorf("Matched %v out of %v signatures", iter, len(expect))
	}
}

func TestTip(t *testing.T) {
	expect := "fmt/669"
	err := setup()
	if err != nil {
		t.Error(err)
	}
	buf := bytes.NewReader([]byte{0x00, 0x4d, 0x52, 0x4d, 0x00})
	c, err := s.Identify("test.mrw", buf)
	for i := range c {
		if i.String() != expect {
			t.Errorf("First buffer: expecting %s, got %s", expect, i)
		}
	}
	buf = bytes.NewReader([]byte{0x00, 0x4d, 0x52, 0x4d, 0x00})
	c, err = s.Identify("test.mrw", buf)
	for i := range c {
		if i.String() != expect {
			t.Errorf("Second buffer: expecting %s, got %s", expect, i)
		}
	}
	buf = bytes.NewReader([]byte{0x00, 0x4d, 0x52, 0x4d, 0x00})
	c, err = s.Identify("test.mrw", buf)
	for i := range c {
		if i.String() != expect {
			t.Errorf("Third buffer: expecting %s, got %s", expect, i)
		}
	}
}

func Test363(t *testing.T) {
	repetitions := 10000
	iter := 0
	expect := "fmt/363"
	err := setup()
	if err != nil {
		t.Error(err)
	}
	segy := func(l int) []byte {
		b := make([]byte, l)
		for i := range b {
			if i > 21 {
				break
			}
			b[i] = 64
		}
		copy(b[l-9:], []byte{01, 00, 00, 00, 01, 00, 00, 01, 00})
		return b
	}
	se := segy(3226)
	for i := 0; i < repetitions; i++ {
		buf := bytes.NewReader(se)
		c, _ := s.Identify("test.seg", buf)
		for i := range c {
			iter++
			if i.String() != expect {
				sbuf := s.Buffer()
				equal := true
				if !bytes.Equal(se, siegreader.Bytes(sbuf)) {
					equal = false
				}
				t.Errorf("First buffer on %d iteration: expecting %s, got %s, buffer equality test is %v", iter, expect, i, equal)
			}
		}
	}
	iter = 0
	se = segy(3626)
	for i := 0; i < repetitions; i++ {
		buf := bytes.NewReader(se)
		c, _ := s.Identify("test2.seg", buf)
		for i := range c {
			iter++
			if i.String() != expect {
				sbuf := s.Buffer()
				equal := true
				if !bytes.Equal(se, siegreader.Bytes(sbuf)) {
					equal = false
				}
				t.Errorf("Second buffer on %d iteration: expecting %s, got %s, buffer equality test is %v", iter, expect, i, equal)
			}
		}
	}
}

// Benchmarks
func BenchmarkNew(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		setup(config.SetDoubleUp())
	}
}

func benchidentify(ext string) {
	file := filepath.Join(*testdata, "benchmark", "Benchmark")
	file += "." + ext
	identifyT(s, file)
}

func BenchmarkACCDB(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("accdb")
	}
}

func BenchmarkBMP(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("bmp")
	}
}

func BenchmarkDOCX(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("docx")
	}
}

func BenchmarkGIF(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("gif")
	}
}

func BenchmarkJPG(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("jpg")
	}
}

func BenchmarkMSG(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("msg")
	}
}

func BenchmarkODT(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("odt")
	}
}

func BenchmarkPDF(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("pdf")
	}
}

func BenchmarkPNG(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("png")
	}
}

func BenchmarkPPTX(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("pptx")
	}
}

func BenchmarkRTF(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("rtf")
	}
}

func BenchmarkTIF(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("tif")
	}
}

func BenchmarkXLSX(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("xlsx")
	}
}

func BenchmarkXML(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		benchidentify("xml")
	}
}

func BenchmarkMulti(bench *testing.B) {
	dir := filepath.Join(*testdata, "benchmark")
	for i := 0; i < bench.N; i++ {
		multiIdentifyT(s, dir)
	}
}
