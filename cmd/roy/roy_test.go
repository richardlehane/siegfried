package main

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/loc"
	"github.com/richardlehane/siegfried/mimeinfo"
	"github.com/richardlehane/siegfried/pronom"
)

var testhome = flag.String("testhome", "data", "override the default home directory")

func TestMakeDefault(t *testing.T) {
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

func TestMakePronomTika(t *testing.T) {
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
	m, err := mimeinfo.New(config.SetName("tika"), config.SetMIMEInfo("tika"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(m)
	if err != nil {
		t.Fatal(err)
	}
	l, err := loc.New(config.SetName("loc"), config.SetLOC(""))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Add(l)
	if err != nil {
		t.Fatal(err)
	}
	sigs := filepath.Join("data", "pronom-tika-loc.sig")
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
