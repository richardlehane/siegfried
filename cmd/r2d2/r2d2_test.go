package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
)

var testhome = flag.String("testhome", "data", "override the default home directory")

func TestMakeGob(t *testing.T) {
	config.Siegfried.Home = *testhome
	if err := config.SetLatest(); err != nil {
		t.Fatal(err)
	}
	s := siegfried.New()
	err := s.Add(siegfried.Pronom)
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
