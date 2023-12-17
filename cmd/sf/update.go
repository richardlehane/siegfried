// Copyright 2015 Richard Lehane. All rights reserved.
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
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/pkg/config"
)

type Updates []Update

type Update struct {
	Version [3]int `json:"sf"`
	Created string `json:"created"`
	Hash    string `json:"hash"`
	Size    int    `json:"size"`
	Path    string `json:"path"`
}

func current(buf []byte, utime string) bool {
	ut, err := time.Parse(time.RFC3339, utime)
	if err != nil {
		return false
	}
	if len(buf) < len(config.Magic())+2+15 {
		return false
	}
	rc := flate.NewReader(bytes.NewBuffer(buf[len(config.Magic())+2:]))
	nbuf := make([]byte, 15)
	if n, _ := rc.Read(nbuf); n < 15 {
		return false
	}
	rc.Close()
	ls := persist.NewLoadSaver(nbuf)
	tt := ls.LoadTime()
	if ls.Err != nil {
		return false
	}
	return !ut.After(tt)
}

func same(buf []byte, usize int, uhash string) bool {
	if len(buf) != usize {
		return false
	}
	h := sha256.New()
	h.Write(buf)
	return hex.EncodeToString(h.Sum(nil)) == uhash
}

func uptodate(utime, uhash string, usize int) bool {
	fbuf, err := os.ReadFile(config.Signature())
	if err != nil {
		return false
	}
	if current(fbuf, utime) && same(fbuf, usize, uhash) {
		return true
	}
	return false
}

func location(base, sig string, args []string) string {
	if len(args) > 0 && len(args[0]) > 0 {
		if args[0] == "freedesktop.org" { // freedesktop.org is more correct, but we don't use it in update service
			args[0] = "freedesktop"
		}
		return base + "/" + args[0]
	}
	if len(sig) > 0 {
		return base + "/" + strings.TrimSuffix(filepath.Base(sig), filepath.Ext(sig))
	}
	return base
}

func updateSigs(sig string, args []string) (bool, string, error) {
	return updateSigsDo(sig, args, getHttp)
}

func updateSigsDo(sig string, args []string, gf getHttpFn) (bool, string, error) {
	url, _, _ := config.UpdateOptions()
	if url == "" {
		return false, "Update is not available for this distribution of siegfried", nil
	}
	response, err := gf(location(url, sig, args))
	if err != nil {
		return false, "", err
	}
	var us Updates
	if err := json.Unmarshal(response, &us); err != nil {
		return false, "", err
	}
	version := config.Version()
	var u Update
	for _, v := range us {
		if version[0] == v.Version[0] && version[1] == v.Version[1] {
			u = v
			break
		}
	}
	if u.Version == [3]int{0, 0, 0} { // we didn't find an eligible update
		return false, "Your version of siegfried is out of date; please install latest from http://www.itforarchivists.com/siegfried before continuing.", nil
	}
	if uptodate(u.Created, u.Hash, u.Size) {
		return false, "You are already up to date!", nil
	}
	// this hairy bit of golang exception handling is thanks to Ross! :)
	if _, err = os.Stat(config.Home()); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(config.Home(), os.ModePerm)
			if err != nil {
				return false, "", fmt.Errorf("siegfried: cannot create home directory %s, %v", config.Home(), err)
			}
		} else {
			return false, "", fmt.Errorf("siegfried: error opening directory %s, %v", config.Home(), err)
		}
	}
	fmt.Println("... downloading latest signature file ...")
	response, err = gf(u.Path)
	if err != nil {
		return false, "", fmt.Errorf("siegfried: retrieving %s.\nThis may be a network or firewall issue. See https://github.com/richardlehane/siegfried/wiki/Getting-started for manual instructions.\nSystem error: %v", config.SignatureBase(), err)
	}
	if !same(response, u.Size, u.Hash) {
		return false, "", fmt.Errorf("siegfried: retrieving %s; SHA256 hash of response doesn't match %s", config.SignatureBase(), u.Hash)
	}
	err = os.WriteFile(config.Signature(), response, os.ModePerm)
	if err != nil {
		return false, "", fmt.Errorf("siegfried: error writing to directory, %v", err)
	}
	fmt.Printf("... writing %s ...\n", config.Signature())
	return true, "Your signature file has been updated", nil
}

type getHttpFn func(string) ([]byte, error)

func getHttp(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	_, timeout, transport := config.UpdateOptions()
	req.Header.Add("User-Agent", config.UserAgent())
	req.Header.Add("Cache-Control", "no-cache")
	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
