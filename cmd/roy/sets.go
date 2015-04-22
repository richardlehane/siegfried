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
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

func replaceKeys(l string) string {
	uniqs := make(map[string]struct{})
	items := strings.Split(l, ",")
	for _, v := range items {
		item := strings.TrimSpace(v)
		if strings.HasPrefix(item, "@") {
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
	if sets == nil {
		if err := initSets(); err != nil {
			return nil, err
		}
	}
	return sets["a"], nil

	// 2. expand any keys into lists of fmts
	// 3. recursively expand keys within those lists. Prevent cycles by adding keys already expanded to a stop list
	// 4. get a comma separted list as input, return a comma separated list as output
}

func initSets() error {
	// 1. load all files in the sets directory and, if no yaml error, add them to a single map
	sets = make(map[string][]string)
	return yaml.Unmarshal([]byte{1, 2}, &sets)
}
