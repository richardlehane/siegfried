package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	sigfile string

	defaultSigs = "pronom.gob"
)

func init() {
	current, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	defaultSigs = filepath.Join(current.HomeDir, "siegfried", defaultSigs)

	flag.StringVar(&sigfile, "sigs", defaultSigs, "path to Siegfried signature file")
}

func load(sigs string) (*core.Siegfried, error) {
	s := core.NewSiegfried()
	p, err := pronom.Load(sigs)
	if err != nil {
		return nil, err
	}
	s.AddIdentifier(p)
	return s, nil
}

func identify(s *core.Siegfried, p string) ([]string, error) {
	ids := make([]string, 0)
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	c, err := s.Identify(file)
	if err != nil {
		return nil, fmt.Errorf("Error with file %v; error: %v", p, err)
	}
	for i := range c {
		ids = append(ids, i.String())
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func multiIdentify(s *core.Siegfried, r string) ([][]string, error) {
	set := make([][]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		ids, err := identify(s, path)
		if err != nil {
			return err
		}
		set = append(set, ids)
		return nil
	}
	err := filepath.Walk(r, wf)
	return set, err
}

func multiIdentifyP(s *core.Siegfried, r string) error {
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		c, err := s.Identify(file)
		if err != nil {
			return err
		}
		fmt.Println(path)
		for i := range c {
			fmt.Println(i)
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
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	s, err := load(sigfile)
	if err != nil {
		log.Fatal(err)

	}

	if info.IsDir() {
		file.Close()
		err = multiIdentifyP(s, flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	c, err := s.Identify(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(flag.Arg(0))
	for i := range c {
		fmt.Println(i)
	}
	file.Close()

	os.Exit(0)
}
