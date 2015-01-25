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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"

	//_ "net/http/pprof"
)

const (
	CONCURRENT = 4
	PROCS      = -1
)

var (
	update  = flag.Bool("update", false, "update or install the default signature file")
	version = flag.Bool("version", false, "display version information")
	debug   = flag.Bool("debug", false, "scan in debug mode")
	nr      = flag.Bool("nr", false, "prevent automatic directory recursion")
	csvo    = flag.Bool("csv", false, "CSV output format")
	sig     = flag.String("sig", config.SignatureBase(), "set the signature file")
	home    = flag.String("home", config.Home(), "override the default home directory")
	serve   = flag.String("serve", "false", "not yet implemented - coming with v1")
)

var (
	csvWriter  *csv.Writer
	yamlWriter *bufio.Writer
)

func getHttp(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	_, timeout, transport := config.UpdateOptions()
	req.Header.Add("User-Agent", "siegfried/siegbot (+https://github.com/richardlehane/siegfried)")
	req.Header.Add("Cache-Control", "no-cache")
	timer := time.AfterFunc(timeout, func() {
		transport.CancelRequest(req)
	})
	defer timer.Stop()
	client := http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type Update struct {
	SfVersion  [3]int
	SigCreated string
	GobSize    int
	LatestURL  string
}

func updateSigs() (string, error) {
	url, _, _ := config.UpdateOptions()
	if url == "" {
		return "Update is not available for this distribution of Siegfried", nil
	}
	response, err := getHttp(url)
	if err != nil {
		return "", err
	}
	var u Update
	if err := json.Unmarshal(response, &u); err != nil {
		return "", err
	}
	version := config.Version()
	if version[0] < u.SfVersion[0] || (u.SfVersion[0] == version[0] && version[1] < u.SfVersion[1]) {
		return "Your version of Siegfried is out of date; please install latest from http://www.itforarchivists.com/siegfried before continuing.", nil
	}
	s, err := siegfried.Load(config.Signature())
	if err == nil {
		if !s.Update(u.SigCreated) {
			return "You are already up to date!", nil
		}
	} else {
		// this hairy bit of golang exception handling is thanks to Ross! :)
		if _, err = os.Stat(config.Home()); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(config.Home(), os.ModePerm)
				if err != nil {
					return "", fmt.Errorf("Siegfried: cannot create home directory %s, %v", config.Home(), err)
				}
			} else {
				return "", fmt.Errorf("Siegfried: error opening directory %s, %v", config.Home(), err)
			}
		}
	}
	fmt.Println("... downloading latest signature file ...")
	response, err = getHttp(u.LatestURL)
	if err != nil {
		return "", err
	}
	if len(response) != u.GobSize {
		return "", fmt.Errorf("Siegfried: error retrieving pronom.gob; expecting %d bytes, got %d bytes", u.GobSize, len(response))
	}
	err = ioutil.WriteFile(config.Signature(), response, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Siegfried: error writing to directory, %v", err)
	}
	fmt.Printf("... writing %s ...\n", config.Signature())
	return "Your signature file has been updated", nil
}

func load() (*siegfried.Siegfried, error) {
	s, err := siegfried.Load(config.Signature())
	if err != nil {
		return nil, err
	}
	return s, nil
}

func identify(s *siegfried.Siegfried, p string) ([]string, error) {
	ids := make([]string, 0)
	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open %v, got: %v", p, err)
	}
	c, err := s.Identify(p, file)
	if c == nil {
		return nil, fmt.Errorf("failed to identify %v, got: %v", p, err)
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

func multiIdentify(s *siegfried.Siegfried, r string) ([][]string, error) {
	set := make([][]string, 0)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *nr && path != r {
				return filepath.SkipDir
			}
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

/*
var lastPath string

func quitter() {
	timer := time.NewTimer(time.Minute * 25)
	<-timer.C
	panic(lastPath)
}
*/
func multiIdentifyP(s *siegfried.Siegfried, r string) {
	//go quitter()
	wg := &sync.WaitGroup{}
	runtime.GOMAXPROCS(PROCS)
	resc := make(chan chan res, CONCURRENT)
	go printer(resc, wg)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if *nr && path != r {
				return filepath.SkipDir
			}
			return nil
		}
		//lastPath = path
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
			rchan <- res{path, info.Size(), ids, err}
			file.Close()
		}()
		return nil
	}
	filepath.Walk(r, wf)
	wg.Wait()
	close(resc)
}

func PrintFile(name string, sz int64, err error) {
	fmt.Print(fileString(name, sz, err))
}

func fileString(name string, sz int64, err error) string {
	var errStr string
	if err != nil {
		errStr = fmt.Sprintf("\"%s\"", err.Error())
	}
	return fmt.Sprintf("---\nfilename : \"%s\"\nfilesize : %d\nerrors   : %s\nmatches  :\n", name, sz, errStr)
}

func main() {
	/*
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	*/
	flag.Parse()

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
	s, err := load()
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
		multiIdentifyP(s, flag.Arg(0))
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
		PrintFile(flag.Arg(0), info.Size(), err)
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
