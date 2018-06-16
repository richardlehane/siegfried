package rk

import (
	"bytes"
	"testing"
)

func equal(a []int, b []Result) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i].Index {
			return false
		}
	}
	return true
}

func toStrings(b [][]byte) (s []string) {
	s = make([]string, len(b))
	for i, v := range b {
		s[i] = string(v)
	}
	return
}

func toBytes(s ...string) (bytes [][]byte) {
	bytes = make([][]byte, len(s))
	for i, v := range s {
		bytes[i] = []byte(v)
	}
	return
}

func indexes(a, b [][]byte) []int {
	indexes := make([]int, 0)
	for _, v := range b {
		for i, v1 := range a {
			if bytes.Equal(v, v1) {
				indexes = append(indexes, i)
			}
		}
	}
	return indexes
}

func loop(output chan Result) []Result {
	results := make([]Result, 0)
	for res := range output {
		results = append(results, res)
	}
	return results
}

func tester(t *testing.T, rk *Rk, a []byte, b, c [][]byte) {
	i := indexes(b, c)
	output := rk.Index(bytes.NewBuffer(a))
	results := loop(output)
	if !equal(i, results) {
		t.Errorf("Index fail; Expecting: %v, Got: %v", i, results)
	}
}

func test(t *testing.T, a []byte, b, c [][]byte) {
	rk, _ := New(b)
	tester(t, rk, a, b, c)
}

func noResult() [][]byte {
	return make([][]byte, 0)
}

// Tests (the test strings are taken from John Graham-Cumming's lua implementation: https://github.com/jgrahamc/aho-corasick-lua Copyright (c) 2013 CloudFlare)
func TestWikipedia(t *testing.T) {
	test(t, []byte("abccab"), toBytes("ab", "bc", "ca"), toBytes("ab", "bc", "ca", "ab"))
}

func TestSimple(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("poto"), noResult())
	test(t, []byte("The pot had a handle"), toBytes("The"), toBytes("The"))
	test(t, []byte("The pot had a handle"), toBytes("pot"), toBytes("pot"))
	test(t, []byte("The pot had a handle"), toBytes("pot "), toBytes("pot "))
	test(t, []byte("The pot had a handle"), toBytes("ot h"), toBytes("ot h"))
	test(t, []byte("The pot had a handle"), toBytes("andle"), toBytes("andle"))
}

func TestMultipleNonoverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("h"), toBytes("h", "h", "h"))
	test(t, []byte("The pot had a handle"), toBytes("ha", "he"), toBytes("he", "ha", "ha"))
	test(t, []byte("The pot had a handle"), toBytes("pot", "had"), toBytes("pot", "had"))
	test(t, []byte("The pot had a handle"), toBytes("pot", "had", "hod"), toBytes("pot", "had"))
	test(t, []byte("The pot had a handle"), toBytes("The", "pot", "had", "hod", "and"),
		toBytes("The", "pot", "had", "and"))
}

func TestOverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("e pot", "pot h"),
		toBytes("e pot", "pot h"))
}

func TestNesting(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("handl", "andle"),
		toBytes("handl", "andle"))
	test(t, []byte("The pot had a handle"), toBytes("he ", "pot", "dle", "xyz"),
		toBytes("he ", "pot", "dle"))
}

func TestRandom(t *testing.T) {
	test(t, []byte("yasherhs"), toBytes("say", "she", "shr", "her"),
		toBytes("she", "her"))
}

func TestFailPartial(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("dlf", "had"), toBytes("had"))
}

func TestMany(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("e", "a", "n", "d", "l"),
		toBytes("e", "a", "d", "a", "a", "n", "d", "l", "e"))
}

func TestLong(t *testing.T) {
	test(t, []byte("macintosh"), toBytes("acintosh"), toBytes("acintosh"))
}

func TestWindow(t *testing.T) {
	win := new(window)
	for i := 0; i < 60; i++ {
		win.push('a')
	}
	if !win.equals([]byte("aaaaa")) {
		t.Error("Window fail, aaaaa should match")
	}
	win.push('b')
	if !win.equals([]byte("aaaab")) {
		t.Error("Window fail, aaaab should match")
	}
	win.push('c')
	win.push('d')
	win.push('e')
	if !win.equals([]byte("abcde")) {
		t.Error("Window fail, abcde should match")
	}
	win.push('f')
	if !win.equals([]byte("abcdef")) {
		t.Error("Window fail, abcdef should match")
	}
	win.push('g')
	if !win.equals([]byte("efg")) {
		t.Error("Window fail, efg should match")
	}
}

// Benchmarks
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(toBytes("handle", "andley", "candle", "person", "apples", "fruits"))
	}
}

func BenchmarkIndex(b *testing.B) {
	b.StopTimer()
	rk, _ := New(toBytes("handle", "andley", "candle", "person", "apples", "fruits"))
	input := bytes.NewBuffer([]byte("The pot had a handle"))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range rk.Index(input) {
		}
	}
}
