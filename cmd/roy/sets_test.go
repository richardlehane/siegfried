package main

import (
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/pkg/config"
)

func TestSets(t *testing.T) {
	config.SetHome(*testhome)
	list := "fmt/1,fmt/2,@pdfa,x-fmt/19"
	expect := "fmt/1,fmt/2,fmt/95,fmt/354,fmt/476,fmt/477,fmt/478,fmt/479,fmt/480,fmt/481,x-fmt/19"
	res := strings.Join(expandSets(list), ",")
	if res != expect {
		t.Errorf("expecting %s, got %s", expect, res)
	}
	pdfs := strings.Join(expandSets("@pdf"), ",")
	expect = "fmt/14,fmt/15,fmt/16,fmt/17,fmt/18,fmt/19,fmt/20,fmt/95,fmt/144,fmt/145,fmt/146,fmt/147,fmt/148,fmt/157,fmt/158,fmt/276,fmt/354,fmt/476,fmt/477,fmt/478,fmt/479,fmt/480,fmt/481,fmt/488,fmt/489,fmt/490,fmt/491,fmt/492,fmt/493"
	if pdfs != expect {
		t.Errorf("expecting %s, got %s", expect, pdfs)
	}
	compression := strings.Join(expandSets("@compression"), ",")
	expect = "fmt/626,x-fmt/266,x-fmt/267,x-fmt/268"
	if compression != expect {
		t.Errorf("expecting %s, got %s", expect, compression)
	}
}

var testSet = map[string][]string{
	"t": {"a", "a", "b", "c"},
	"u": {"b", "d"},
	"v": {"@t", "@u"},
}

func TestDupeSets(t *testing.T) {
	sets = testSet
	expect := "a,b,c,d"
	res := strings.Join(expandSets("@v"), ",")
	if res != expect {
		t.Errorf("expecting %s, got %s", expect, res)
	}
}
