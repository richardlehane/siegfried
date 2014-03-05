package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var testSigs = filepath.Join("..", "r2d2", "data", "pronom.gob")

var s *core.Siegfried

func init() {
	var err error
	s, err = load(testSigs)
	if err != nil {
		panic(err)
	}
}

var root = filepath.Join(".", "testdata", "skeleton-suite")

func TestLoad(t *testing.T) {
	_, err := load(testSigs)
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
	expect := make([]string, 0)
	names := make([]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		last := strings.Split(path, string(os.PathSeparator))
		path = last[len(last)-1]
		idx := strings.Index(path, "signature")
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
	err := filepath.Walk(root, wf)
	if err != nil {
		t.Fatal(err)
	}
	matches, err := multiIdentify(b, root)
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
		load(testSigs)
	}
}

func benchidentify(ext string) {
	file := filepath.Join(".", "testdata", "benchmark", "Benchmark")
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
