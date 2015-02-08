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
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"

	"net/http"
	_ "net/http/pprof"
)

const PROCS = -1

// flags
var (
	update  = flag.Bool("update", false, "update or install the default signature file")
	version = flag.Bool("version", false, "display version information")
	debug   = flag.Bool("debug", false, "scan in debug mode")
	nr      = flag.Bool("nr", false, "prevent automatic directory recursion")
	csvo    = flag.Bool("csv", false, "CSV output format")
	sig     = flag.String("sig", config.SignatureBase(), "set the signature file")
	home    = flag.String("home", config.Home(), "override the default home directory")
	serve   = flag.String("serve", "false", "not yet implemented - coming with v1")
	multi   = flag.Int("multi", 1, "set number of file ID processes")
	profile = flag.Bool("profile", false, "run a profile on localhost:6060")
)

var (
	csvWriter  *csv.Writer
	yamlWriter *bufio.Writer
	replacer   = strings.NewReplacer("'", "''")
)

type res struct {
	path string
	sz   int64
	c    []core.Identification
	err  error
}

func printer(resc chan chan res, wg *sync.WaitGroup) {
	var csvRecord []string
	if *csvo {
		csvRecord = make([]string, 10)
	}
	for rr := range resc {
		r := <-rr
		if !config.Debug() && !*csvo {
			yamlWriter.WriteString(fileString(r.path, r.sz, r.err))
		}
		if r.c == nil {
			if *csvo {
				var errStr string
				if r.err != nil {
					errStr = r.err.Error()
				}
				empty := make([]string, 7)
				csvRecord[0], csvRecord[1], csvRecord[2] = r.path, strconv.Itoa(int(r.sz)), errStr
				copy(csvRecord[3:], empty)
				csvWriter.Write(csvRecord)
			}
			continue
		}
		for _, v := range r.c {
			switch {
			case config.Debug():
			case *csvo:
				var errStr string
				if r.err != nil {
					errStr = r.err.Error()
				}
				csvRecord[0], csvRecord[1], csvRecord[2] = r.path, strconv.Itoa(int(r.sz)), errStr
				copy(csvRecord[3:], v.Csv())
				csvWriter.Write(csvRecord)
			default:
				yamlWriter.WriteString(v.Yaml())
			}
		}
		wg.Done()
	}
}

func multiIdentifyP(s *siegfried.Siegfried, r string) {
	wg := &sync.WaitGroup{}
	runtime.GOMAXPROCS(PROCS)
	resc := make(chan chan res, *multi)
	go printer(resc, wg)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *nr && path != r {
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

func multiIdentifyS(s *siegfried.Siegfried, r string) error {
	var csvRecord []string
	if *csvo {
		csvRecord = make([]string, 10)
	}
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *nr && path != r {
				return filepath.SkipDir
			}
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			if !*csvo {
				yamlWriter.WriteString(fileString(path, info.Size(), fmt.Errorf("failed to open %s, got: %v", path, err)))
				return nil
			}
			empty := make([]string, 7)
			csvRecord[0], csvRecord[1], csvRecord[2] = path, strconv.Itoa(int(info.Size())), fmt.Sprintf("failed to open %s, got: %v", path, err)
			copy(csvRecord[3:], empty)
			csvWriter.Write(csvRecord)
			return nil
		}
		c, err := s.Identify(path, file)
		if c == nil {
			file.Close()
			if !*csvo {
				yamlWriter.WriteString(fileString(path, info.Size(), fmt.Errorf("failed to identify %s, got: %v", path, err)))
				return nil
			}
			empty := make([]string, 7)
			csvRecord[0], csvRecord[1], csvRecord[2] = path, strconv.Itoa(int(info.Size())), fmt.Sprintf("failed to identify %s, got: %v", path, err)
			copy(csvRecord[3:], empty)
			csvWriter.Write(csvRecord)
			return nil
		}
		if !config.Debug() && !*csvo {
			yamlWriter.WriteString(fileString(path, info.Size(), err))
		}
		for i := range c {
			switch {
			case config.Debug():
			case *csvo:
				var errStr string
				if err != nil {
					errStr = err.Error()
				}
				csvRecord[0], csvRecord[1], csvRecord[2] = path, strconv.Itoa(int(info.Size())), errStr
				copy(csvRecord[3:], i.Csv())
				csvWriter.Write(csvRecord)
			default:
				yamlWriter.WriteString(i.Yaml())
			}
		}
		return file.Close() // this op is unreasonably slow on OSX
	}
	return filepath.Walk(r, wf)
}

func fileString(name string, sz int64, err error) string {
	var errStr string
	if err != nil {
		errStr = fmt.Sprintf("\"%s\"", err.Error())
	}
	return fmt.Sprintf("---\nfilename : '%s'\nfilesize : %d\nerrors   : %s\nmatches  :\n", replacer.Replace(name), sz, errStr)
}

func main() {

	flag.Parse()
	if *profile {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	if *csvo {
		csvWriter = csv.NewWriter(os.Stdout)
		csvWriter.Write([]string{"filename", "filesize", "errors", "identifier", "id", "format name", "format version", "mimetype", "basis", "warning"})
	} else {
		yamlWriter = bufio.NewWriter(os.Stdout)
	}

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

	if *debug {
		config.SetDebug()
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
		fmt.Println("sf server not yet implemented; expect by v1")
	}

	if flag.NArg() != 1 {
		log.Fatal("Error: expecting a single file or directory argument")
	}

	var err error
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error: error opening %v, got: %v", flag.Arg(0), err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Error: error getting info for %v, got: %v", flag.Arg(0), err)
	}
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		log.Fatalf("Error: error loading signature file, got: %v", err)

	}
	if info.IsDir() {
		file.Close()
		if config.Debug() {
			log.Fatalln("Error: when scanning in debug mode, give a file rather than a directory argument")
		}
		if !*csvo {
			yamlWriter.WriteString(s.Yaml())
		}
		if *multi > 16 {
			*multi = 16
		}
		if *multi > 1 {
			multiIdentifyP(s, flag.Arg(0))
		} else {
			multiIdentifyS(s, flag.Arg(0))
		}
		if *csvo {
			csvWriter.Flush()
		} else {
			yamlWriter.Flush()
		}
		os.Exit(0)
	}
	c, err := s.Identify(flag.Arg(0), file)
	if c == nil {
		file.Close()
		log.Fatal(err)
	}
	if !config.Debug() && !*csvo {
		fmt.Print(s.Yaml())
		fmt.Print(fileString(flag.Arg(0), info.Size(), err))
	}
	var csvRecord []string
	if *csvo {
		csvRecord = make([]string, 10)
	}
	for i := range c {
		switch {
		case config.Debug():
		case *csvo:
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			csvRecord[0], csvRecord[1], csvRecord[2] = flag.Arg(0), strconv.Itoa(int(info.Size())), errStr
			copy(csvRecord[3:], i.Csv())
			csvWriter.Write(csvRecord)
		default:
			fmt.Print(i.Yaml())
		}
	}

	file.Close()
	if *csvo {
		csvWriter.Flush()
	}

	os.Exit(0)
}
