package main

import (
	"strings"
	"testing"

	"github.com/richardlehane/siegfried/config"
)

func TestSets(t *testing.T) {
	config.SetHome(*testhome)
	list := "fmt/1,fmt/2,@pdfa,x-fmt/19"
	expect := "fmt/1,fmt/2,fmt/95,fmt/354,fmt/476,fmt/477,fmt/478,fmt/479,fmt/480,fmt/481,x-fmt/19"
	res := strings.Join(expandSets(list), ",")
	if res != expect {
		t.Errorf("expecting %s, got %s", expect, res)
	}
}
