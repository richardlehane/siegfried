package main

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var testhome = flag.String("testhome", "data", "override the default home directory")

func TestMakeGob(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New()
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	sigs := filepath.Join("data", config.SignatureBase())
	err = s.Save(sigs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMakeArchivematica(t *testing.T) {
	s := siegfried.New()
	config.SetHome(*testhome)
	p, err := pronom.New(
		config.SetName("archivematica"),
		config.SetExtend(expandSets("archivematica-fmt2.xml,archivematica-fmt3.xml,archivematica-fmt4.xml,archivematica-fmt5.xml")))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(p)
	if err != nil {
		t.Fatal(err)
	}
	sigs := filepath.Join("data", "archivematica.sig")
	err = s.Save(sigs)
	if err != nil {
		t.Fatal(err)
	}
}
