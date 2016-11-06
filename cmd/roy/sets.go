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
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/config"
)

// take a comma separated string of puids and sets (e.g. fmt/1,@pdf,fmt/2) and expand any sets within.
// Also split, trim, sort and de-dupe.
// return a slice of puids
func expandSets(l string) []string {
	uniqs := make(map[string]struct{}) // drop any duplicates with this map
	items := strings.Split(l, ",")
	for _, v := range items {
		item := strings.TrimSpace(v)
		if strings.HasPrefix(item, "@") {
			if sets == nil {
				if err := initSets(); err != nil {
					log.Fatalf("error loading sets: %v", err)
				}
			}
			list, err := getSets(strings.TrimPrefix(item, "@"))
			if err != nil {
				log.Fatalf("error interpreting sets: %v", err)
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
	return sortFmts(ret)
}

// a plain string sort doesn't work e.g. get fmt/1,fmt/111/fmt/2 - need to sort on ints
func sortFmts(s []string) []string {
	fmts := make(map[string][]int)
	others := []string{}
	addFmt := func(str, prefix string) bool {
		if strings.HasPrefix(str, prefix+"/") {
			no, err := strconv.Atoi(strings.TrimPrefix(str, prefix+"/"))
			if err == nil {
				fmts[prefix] = append(fmts[prefix], no)
			} else {
				others = append(others, str)
			}
			return true
		}
		return false
	}
	for _, v := range s {
		if !addFmt(v, "fmt") {
			if !addFmt(v, "x-fmt") {
				others = append(others, v)
			}
		}
	}
	var ret []string
	appendFmts := func(prefix string) {
		f, ok := fmts[prefix]
		if ok {
			sort.Ints(f)
			for _, i := range f {
				ret = append(ret, prefix+"/"+strconv.Itoa(i))
			}
		}
	}
	appendFmts("fmt")
	appendFmts("x-fmt")
	sort.Strings(others)
	return append(ret, others...)
}

var sets map[string][]string

func getSets(key string) ([]string, error) {
	// recursively build a list of all values for the key
	attempted := make(map[string]bool) // prevent cycles by bookkeeping with attempted map
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
		var nl []string
		for _, k2 := range l {
			if strings.HasPrefix(k2, "@") {
				l2, err := f(strings.TrimPrefix(k2, "@"))
				if err != nil {
					return nil, err
				}
				nl = append(nl, l2...)
			} else {
				nl = append(nl, k2)
			}
		}
		return nl, nil
	}
	return f(key)
}

func stripComment(in string) string {
	ws := strings.Index(in, " ")
	if ws < 0 {
		return in
	} else {
		return in[:ws]
	}
}

func stripComments(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = stripComment(v)
	}
	return out
}

func initSets() error {
	//  load all json files in the sets directory and add them to a single map
	sets = make(map[string][]string)
	wf := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("error walking sets directory, must have a 'sets' directory in siegfried home: " + err.Error())
		}
		if info.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		default:
			return nil // ignore non json files
		case ".json":
		}
		set := make(map[string][]string)
		byts, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.New("error loading " + path + " " + err.Error())
		}
		err = json.Unmarshal(byts, &set)
		if err != nil {
			return errors.New("error unmarshalling " + path + " " + err.Error())
		}
		for k, v := range set {
			k = stripComment(k)
			v = stripComments(v)
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
