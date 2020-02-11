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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/internal/persist"
	"github.com/richardlehane/siegfried/pkg/config"
)

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
	fbuf, err := ioutil.ReadFile(config.Signature())
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

func updateSigs(sig string, args []string) (string, error) {
	url, _, _ := config.UpdateOptions()
	if url == "" {
		return "Update is not available for this distribution of siegfried", nil
	}
	response, err := getHttp(location(url, sig, args))
	if err != nil {
		return "", err
	}
	var u Update
	if err := json.Unmarshal(response, &u); err != nil {
		return "", err
	}
	version := config.Version()
	if version[0] < u.Version[0] || (version[0] == u.Version[0] && version[1] < u.Version[1]) || // if the version is out of date
		u.Version == [3]int{0, 0, 0} || u.Created == "" || u.Size == 0 || u.Path == "" { // or if the unmarshalling hasn't worked and we have blank values
		return "Your version of siegfried is out of date; please install latest from http://www.itforarchivists.com/siegfried before continuing.", nil
	}
	if uptodate(u.Created, u.Hash, u.Size) {
		return "You are already up to date!", nil
	}
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
	fmt.Println("... downloading latest signature file ...")
	response, err = getHttp(u.Path)
	if err != nil {
		return "", fmt.Errorf("Siegfried: error retrieving %s.\nThis may be a network or firewall issue. See https://github.com/richardlehane/siegfried/wiki/Getting-started for manual instructions.\nSystem error: %v", config.SignatureBase(), err)
	}
	if !same(response, u.Size, u.Hash) {
		return "", fmt.Errorf("Siegfried: error retrieving %s; SHA256 hash of response doesn't match %s", config.SignatureBase(), u.Hash)
	}
	err = ioutil.WriteFile(config.Signature(), response, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Siegfried: error writing to directory, %v", err)
	}
	fmt.Printf("... writing %s ...\n", config.Signature())
	return "Your signature file has been updated", nil
}

func getHttp(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	_, timeout, transport := config.UpdateOptions()
	req.Header.Add("User-Agent", config.UserAgent())
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
