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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/richardlehane/siegfried/config"
)

func replaceKeys(l string) string {
	uniqs := make(map[string]struct{})
	items := strings.Split(l, ",")
	for _, v := range items {
		item := strings.TrimSpace(v)
		if strings.HasPrefix(item, "@") {
			if sets == nil {
				if err := initSets(); err != nil {
					panic(err)
				}
			}
			list, err := getSets(strings.TrimPrefix(item, "@"))
			if err != nil {
				panic(err)
			}
			for _, v := range list {
				uniqs[v] = struct{}{}
			}
		} else {
			uniqs[item] = struct{}{}
		}
	}
	ret := make([]string, 0, len(uniqs))
	for k := range uniqs {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return strings.Join(ret, ",")
}

var sets map[string][]string

func getSets(key string) ([]string, error) {
	// recursively build a list of all values for the key; prevent cycles by bookkeeping with attempted map
	attempted := make(map[string]bool)
	var f func(string) ([]string, error)
	f = func(k string) ([]string, error) {
		if ok := attempted[k]; ok {
			return nil, nil
		}
		attempted[k] = true
		l, ok := sets[k]
		if !ok {
			return nil, errors.New("sets: unknown key " + k)
		}
		for _, k2 := range l {
			if strings.HasPrefix(k2, "@") {
				l2, err := f(strings.TrimPrefix(k2, "@"))
				if err != nil || l2 == nil {
					return nil, err
				}
				l = append(l, l2...)
			}
		}
		return l, nil
	}
	return f(key)
}

func initSets() error {
	//  load all yaml files in the sets directory and add them to a single map
	sets = make(map[string][]string)
	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		default:
			return nil
		case "yml", "yaml":
		}
		set := make(map[string][]string)
		byts, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(byts, &set)
		if err != nil {
			return err
		}
		for k, v := range set {
			sort.Strings(v)
			m, ok := sets[k]
			if !ok {
				sets[k] = v
			} else {
				// if we already have this key, add any new items in its list to the existing list
				for _, w := range v {
					idx := sort.SearchStrings(m, w)
					if idx == len(m) || m[idx] != w {
						m = append(m, w)
					}
				}
				sort.Strings(m)
				sets[k] = m
			}
		}
		return nil
	}
	return filepath.Walk(filepath.Join(config.Home(), "sets"), wf)
}
