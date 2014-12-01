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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
)

var (
	update  = flag.Bool("update", false, "update or install the default signature file")
	version = flag.Bool("version", false, "display version information")
	debug   = flag.Bool("debug", false, "scan in debug mode")
	sig     = flag.String("sig", config.Signature(), "set the signature file")
	home    = flag.String("home", config.Home(), "override the default home directory")
	serve   = flag.String("serve", "false", "not yet implemented - coming with v1")
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
	if err != nil {
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

func multiIdentifyP(s *siegfried.Siegfried, r string) error {
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %v, got: %v", path, err)
		}
		c, err := s.Identify(path, file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to identify %v, got: %v", path, err)
		}
		if !config.Debug() {
			PrintFile(path, info.Size())
		}
		for i := range c {
			if !config.Debug() {
				fmt.Print(i.Yaml())
			}
		}
		file.Close()
		return nil
	}
	return filepath.Walk(r, wf)
}

func PrintFile(name string, sz int64) {
	fmt.Println("---")
	fmt.Printf("filename : \"%v\"\n", name)
	fmt.Printf("filesize : %d\n", sz)
	if !config.Debug() {
		fmt.Print("matches  :\n")
	}
}

func PrintError(err error) {
	fmt.Println("---")
	fmt.Printf("Error : %v", err)
	fmt.Println("---")
}

func main() {

	flag.Parse()

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
		if !config.Debug() {
			fmt.Print(s.Yaml())
		}
		err = multiIdentifyP(s, flag.Arg(0))
		if err != nil {
			PrintError(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	c, err := s.Identify(flag.Arg(0), file)
	if err != nil {
		PrintError(err)
		file.Close()
		os.Exit(1)
	}
	if !config.Debug() {
		fmt.Print(s.Yaml())
		PrintFile(flag.Arg(0), info.Size())
	}
	for i := range c {
		if !config.Debug() {
			fmt.Print(i.Yaml())
		}

	}
	file.Close()

	os.Exit(0)
}
