package rkac

import (
	"bytes"
	"fmt"
	"testing"
)

func equal(a []int, b []Result) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		for j, w := range b {
			if v == w.Index {
				if j+1 == len(b) {
					b = b[:j]
				} else {
					b = append(b[:j], b[j+1:]...)
				}
				break
			}
		}
	}
	if len(b) == 0 {
		return true
	} else {
		return false
	}
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

func test(t *testing.T, a []byte, b, c [][]byte) {
	rkac, _ := New(b)
	i := indexes(b, c)
	output := rkac.Index(bytes.NewBuffer(a))
	results := loop(output)
	if !equal(i, results) {
		t.Errorf("Index fail; Expecting: %v, Got: %v", i, results)
	}
}

func noResult() [][]byte {
	return make([][]byte, 0)
}

// Tests (the test strings are taken from John Graham-Cumming's lua implementation: https://github.com/jgrahamc/aho-corasick-lua Copyright (c) 2013 CloudFlare)
func TestWikipedia(t *testing.T) {
	test(t, []byte("abccab"), toBytes("a", "ab", "bc", "bca", "c", "caa"), toBytes("a", "ab", "bc", "c", "c", "a", "ab"))
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
	test(t, []byte("The pot had a handle"), toBytes("The", "pot", "had", "hod", "andle"),
		toBytes("The", "pot", "had", "andle"))
}

func TestOverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("Th", "he pot", "The", "pot h"),
		toBytes("Th", "The", "he pot", "pot h"))
}

func TestNesting(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("handle", "hand", "and", "andle"),
		toBytes("hand", "and", "handle", "andle"))
	test(t, []byte("The pot had a handle"), toBytes("handle", "hand", "an", "n"),
		toBytes("an", "n", "hand", "handle"))
	test(t, []byte("The pot had a handle"), toBytes("dle", "l", "le"),
		toBytes("l", "dle", "le"))
}

func TestRandom(t *testing.T) {
	test(t, []byte("yasherhs"), toBytes("say", "she", "shr", "he", "her"),
		toBytes("she", "he", "her"))
}

func TestFailPartial(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("dlf", "l"), toBytes("l"))
}

func TestMany(t *testing.T) {
	test(t, []byte("The pot had a handle"), toBytes("handle", "andle", "ndle", "dle", "le", "e"),
		toBytes("e", "handle", "andle", "ndle", "dle", "le", "e"))
	test(t, []byte("The pot had a handle"), toBytes("handle", "handl", "hand", "han", "ha", "a"),
		toBytes("ha", "a", "a", "ha", "a", "han", "hand", "handl", "handle"))
}

func TestLong(t *testing.T) {
	test(t, []byte("mRkacintosh"), toBytes("Rkacintosh", "in"), toBytes("in", "Rkacintosh"))
	test(t, []byte("mRkacintosh"), toBytes("Rkacintosh", "in", "tosh"), toBytes("in", "Rkacintosh", "tosh"))
	test(t, []byte("mRkacintosh"), toBytes("Rkacintosh", "into", "to", "in"),
		toBytes("in", "into", "to", "Rkacintosh"))
}

// Benchmarks
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = New(toBytes("handle", "handl", "hand", "han", "ha", "a"))
	}
}

func BenchmarkIndex(b *testing.B) {
	b.StopTimer()
	rkac, _ := New(toBytes("handle", "handl", "hand", "han", "ha", "a"))
	input := bytes.NewBuffer([]byte("The pot had a handle"))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range rkac.Index(input) {
		}
	}
}

// following benchmark code is from <http://godoc.org/code.google.com/p/ahocorasick> for comparison
func benchmarkValue(n int) []byte {
	input := []byte{}
	for i := 0; i < n; i++ {
		var b byte
		if i%2 == 0 {
			b = 'a'
		} else {
			b = 'b'
		}
		input = append(input, b)
	}
	return input
}

func hardTree() [][]byte {
	ret := [][]byte{}
	str := ""
	for i := 0; i < 2500; i++ {
		// We add a 'q' to the end to make sure we never Rkactually match
		ret = append(ret, []byte(str+string('a'+(i%26))+"q"))
		if i%26 == 25 {
			str = str + string('a'+len(str)%2)
		}
	}
	return ret
}

func BenchmarkMatchingNoMatch(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	rkac, _ := New(toBytes(
		"abababababababd",
		"abababb",
		"abababababq",
	))
	b.StartTimer()
	for _ = range rkac.Index(reader) {
	}
}

func BenchmarkMatchingManyMatches(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	rkac, _ := New(toBytes(
		"ab",
		"ababababababab",
		"ababab",
		"ababababab",
	))
	b.StartTimer()
	for _ = range rkac.Index(reader) {
	}
}

func BenchmarkMatchingHardTree(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	rkac, err := New(hardTree())
	if err == nil {
		b.StartTimer()
		for _ = range rkac.Index(reader) {
		}
	} else {
		fmt.Print(err)
	}
}
