package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var _ = pronom.Droid{} // import for side effects

var (
	defaultSigs    = filepath.Join(".", "pronom.gob")
	defaultDroid   = filepath.Join(".", "DROID_SignatureFile_V72.xml")
	defaultReports = filepath.Join(".", "pronom")
)

var (
	sigfile = flag.String("sigs", defaultSigs, "path to Siegfried signature file")
	droid   = flag.String("droid", defaultDroid, "path to Droid signature file")
	reports = flag.String("reports", defaultReports, "path to Pronom reports directory")
)

var puids []string

func load(sigs string) (*bytematcher.Bytematcher, error) {
	return bytematcher.Load(sigs)
}

func identify(b *bytematcher.Bytematcher, p string) ([]int, error) {
	ids := make([]int, 0)
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	c, err := b.Identify(file)
	if err != nil {
		return nil, fmt.Errorf("Error with file %v; error: %v", p, err)
	}
	for i := range c {
		ids = append(ids, i)
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func multiIdentify(b *bytematcher.Bytematcher, r string) ([][]int, error) {
	set := make([][]int, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		ids, err := identify(b, path)
		if err != nil {
			return err
		}
		set = append(set, ids)
		return nil
	}
	err := filepath.Walk(r, wf)
	return set, err
}

func multiIdentifyP(b *bytematcher.Bytematcher, r string) error {
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		c, err := b.Identify(file)
		if err != nil {
			return err
		}
		fmt.Println(path)
		for i := range c {
			fmt.Println(puids[i])
		}
		fmt.Println()
		file.Close()
		return nil
	}
	return filepath.Walk(r, wf)
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("Error: expecting a single file or directory argument")
	}

	var err error
	puids, err = pronom.PuidsFromDroid(*droid, *reports)

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	b, err := load(*sigfile)

	if info.IsDir() {
		file.Close()
		err = multiIdentifyP(b, flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	c, err := b.Identify(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(flag.Arg(0))
	for i := range c {
		fmt.Println(puids[i])
	}
	file.Close()

	os.Exit(0)
}
