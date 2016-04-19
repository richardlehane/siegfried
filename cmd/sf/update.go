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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
)

type Update struct {
	Version [3]int `json:"sf"`
	Created string `json:"created"`
	Size    int    `json:"size"`
	Path    string `json:"path"`
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
	if version[0] < u.Version[0] || (version[0] == u.Version[0] && version[1] < u.Version[1]) || // if the version is out of date
		u.Version == [3]int{0, 0, 0} || u.Created == "" || u.Size == 0 || u.Path == "" { // or if the unmarshalling hasn't worked and we have blank values
		return "Your version of Siegfried is out of date; please install latest from http://www.itforarchivists.com/siegfried before continuing.", nil
	}
	s, err := siegfried.Load(config.Signature())
	if err == nil {
		if !s.Update(u.Created) {
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
	response, err = getHttp(u.Path)
	if err != nil {
		return "", fmt.Errorf("Siegfried: error retrieving %s.\nThis may be a network or firewall issue. See https://github.com/richardlehane/siegfried/wiki/Getting-started for manual instructions.\nSystem error: %v", config.SignatureBase(), err)
	}
	if len(response) != u.Size {
		return "", fmt.Errorf("Siegfried: error retrieving %s; expecting %d bytes, got %d bytes", config.SignatureBase(), u.Size, len(response))
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
