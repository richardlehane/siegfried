package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var testhome = flag.String("testhome", "data", "override the default home directory")

func TestMakeGob(t *testing.T) {
	s := siegfried.New()
	p, err := pronom.New(config.SetHome(*testhome))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
	sigs := filepath.Join("data", "pronom.gob")
	err = s.Save(sigs)
	if err != nil {
		t.Fatal(err)
	}
}
