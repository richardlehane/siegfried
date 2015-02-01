package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
)

var testhome = flag.String("testhome", filepath.Join("..", "roy", "data"), "override the default home directory")
var testdata = flag.String("testdata", filepath.Join(".", "testdata"), "override the default test data directory")

var s *siegfried.Siegfried

func setup() error {
	var err error
	config.SetHome(*testhome)
	s, err = siegfried.Load(config.Signature())
	return err
}

func identify(s *siegfried.Siegfried, p string) ([]string, error) {
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

func multiIdentify(s *siegfried.Siegfried, r string) ([][]string, error) {
	set := make([][]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *nr && path != r {
				return filepath.SkipDir
			}
			return nil
		}
		ids, err := identify(s, path)
		if err != nil {
			return err
		}
		set = append(set, ids)
		return nil
	}
	err := filepath.Walk(r, wf)
	return set, err
}

func TestLoad(t *testing.T) {
	err := setup()
	if err != nil {
		t.Error(err)
	}
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
	matches, err := multiIdentify(s, suite)
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

// Benchmarks
func BenchmarkNew(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		setup()
	}
}

func benchidentify(ext string) {
	file := filepath.Join(*testdata, "benchmark", "Benchmark")
	file += "." + ext
	identify(s, file)
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
		multiIdentify(s, dir)
	}
}
