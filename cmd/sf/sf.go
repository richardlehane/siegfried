// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"

	// Uncomment to build with profiler
	//"net/http"
	//_ "net/http/pprof"
)

const PROCS = -1

// flags
var (
	update  = flag.Bool("update", false, "update or install the default signature file")
	version = flag.Bool("version", false, "display version information")
	debug   = flag.Bool("debug", false, "scan in debug mode")
	nr      = flag.Bool("nr", false, "prevent automatic directory recursion")
	csvo    = flag.Bool("csv", false, "CSV output format")
	jsono   = flag.Bool("json", false, "JSON output format")
	sig     = flag.String("sig", config.SignatureBase(), "set the signature file")
	home    = flag.String("home", config.Home(), "override the default home directory")
	serve   = flag.String("serve", "false", "not yet implemented - coming with v1")
	multi   = flag.Int("multi", 1, "set number of file ID processes")
	//profile = flag.Bool("profile", false, "run a profile on localhost:6060")
)

type res struct {
	path string
	sz   int64
	c    []core.Identification
	err  error
}

func printer(w writer, resc chan chan res, wg *sync.WaitGroup) {
	for rr := range resc {
		r := <-rr
		w.writeFile(r.path, r.sz, r.err, &idSlice{0, r.c})
		wg.Done()
	}
}

func multiIdentifyP(w writer, s *siegfried.Siegfried, r string, norecurse bool) {
	wg := &sync.WaitGroup{}
	runtime.GOMAXPROCS(PROCS)
	resc := make(chan chan res, *multi)
	go printer(w, resc, wg)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if norecurse && path != r {
				return filepath.SkipDir
			}
			return nil
		}
		wg.Add(1)
		rchan := make(chan res, 1)
		resc <- rchan
		go func() {
			file, err := os.Open(path)
			if err != nil {
				rchan <- res{"", 0, nil, fmt.Errorf("failed to open %v, got: %v", path, err)}
				return
			}
			c, err := s.Identify(path, file)
			if c == nil {
				file.Close()
				rchan <- res{"", 0, nil, fmt.Errorf("failed to identify %v, got: %v", path, err)}
				return
			}
			ids := make([]core.Identification, 0, 1)
			for id := range c {
				ids = append(ids, id)
			}
			cerr := file.Close()
			if err == nil {
				err = cerr
			}
			rchan <- res{path, info.Size(), ids, err}
		}()
		return nil
	}
	filepath.Walk(r, wf)
	wg.Wait()
	close(resc)
}

func multiIdentifyS(w writer, s *siegfried.Siegfried, r string, norecurse bool) error {
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if norecurse && path != r {
				return filepath.SkipDir
			}
			return nil
		}
		identifyFile(w, s, path, info.Size())
		return nil
	}
	return filepath.Walk(r, wf)
}

func identifyFile(w writer, s *siegfried.Siegfried, path string, sz int64) {
	file, err := os.Open(path)
	if err != nil {
		w.writeFile(path, sz, fmt.Errorf("failed to open %s, got: %v", path, err), nil)
		return
	}
	defer file.Close()
	c, err := s.Identify(path, file)
	if c == nil {
		w.writeFile(path, sz, fmt.Errorf("failed to identify %s, got: %v", path, err), nil)
		return
	}
	w.writeFile(path, sz, err, idChan(c))
}

func main() {

	flag.Parse()
	/*
		if *profile {
			go func() {
				log.Println(http.ListenAndServe("localhost:6060", nil))
			}()
		}
	*/

	if *home != config.Home() {
		config.SetHome(*home)
	}

	if *sig != config.SignatureBase() {
		config.SetSignature(*sig)
	}

	if *version {
		version := config.Version()
		fmt.Printf("Siegfried version: %d.%d.%d\n", version[0], version[1], version[2])
		return
	}

	if *update {
		msg, err := updateSigs()
		if err != nil {
			log.Fatalf("Siegfried: error updating signature file, %v", err)
		}
		fmt.Println(msg)
		return
	}

	if *serve != "false" {
		s, err := siegfried.Load(config.Signature())
		if err != nil {
			log.Fatalf("Error: error loading signature file, got: %v", err)

		}
		log.Printf("Starting server at %s. Use CTRL-C to quit.\n", *serve)
		listen(*serve, s)
		return
	}

	if flag.NArg() != 1 {
		log.Fatalln("Error: expecting a single file or directory argument")
	}

	var err error
	info, err := os.Stat(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error: error getting info for %v, got: %v", flag.Arg(0), err)
	}
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		log.Fatalf("Error: error loading signature file, got: %v", err)

	}

	var w writer
	switch {
	case *debug:
		config.SetDebug()
		w = debugWriter{}
	case *csvo:
		w = newCsv(os.Stdout)
	case *jsono:
		w = newJson(os.Stdout)
	default:
		w = newYaml(os.Stdout)
	}

	if info.IsDir() {
		if config.Debug() {
			log.Fatalln("Error: when scanning in debug mode, give a file rather than a directory argument")
		}
		w.writeHead(s)
		if *multi > 16 {
			*multi = 16
		}
		if *multi > 1 {
			multiIdentifyP(w, s, flag.Arg(0), *nr)
		} else {
			multiIdentifyS(w, s, flag.Arg(0), *nr)
		}
		w.writeTail()
		os.Exit(0)
	}

	w.writeHead(s)
	identifyFile(w, s, flag.Arg(0), info.Size())
	w.writeTail()
	os.Exit(0)
}
