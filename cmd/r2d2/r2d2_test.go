package main

import (
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/pkg/pronom"
)

func TestMakeGob(t *testing.T) {
	p, err := pronom.NewIdentifier(pronom.ConfigPaths())
	if err != nil {
		t.Fatal(err)
	}
	sigs := filepath.Join("data", "pronom.gob")
	err = p.Save(sigs)
	if err != nil {
		t.Fatal(err)
	}
}
