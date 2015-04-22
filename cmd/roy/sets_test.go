package main

import (
	"testing"

	"github.com/richardlehane/siegfried/config"
)

func TestSets(t *testing.T) {
	config.SetHome(*testhome)
	list := "fmt/1,fmt/2,@pdf,x-fmt/19"
	expect := "fmt/1,fmt/2,fmt/16,fmt/17,fmt/18,fmt/117,fmt/118,x-fmt/19"
	res := expandSets(list)
	if res != expect {
		t.Errorf("expecting %s, got %s", expect, res)
	}
}
