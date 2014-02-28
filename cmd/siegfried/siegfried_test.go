package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/pkg/pronom"
)

var testSigs = filepath.Join("..", "r2d2", "data", "pronom.gob")

func init() {
	var err error
	droid, _, reports := pronom.ConfigPaths()
	puids, err = pronom.PuidsFromDroid(droid, reports)
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

func check(i string, j []int) bool {
	for _, v := range j {
		if i == puids[v] {
			return true
		}
	}
	return false
}

func matchString(i []int) string {
	str := "[ "
	for _, v := range i {
		str += puids[v]
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
	b, err := load(testSigs)
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
		t.Error(b.Stats())
	}
}
