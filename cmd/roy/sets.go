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

import "gopkg.in/yaml.v2"

func getSets(l string) (string, error) {
	m := make(map[string][]string)
	err := yaml.Unmarshal([]byte{1, 2}, &m)
	if err != nil {
		return nil, err
	}
	return m["a"], nil
	// 1. load all files in the sets directory and, if no yaml error, add them to a single map
	// 2. expand any keys into lists of fmts
	// 3. recursively expand keys within those lists. Prevent cycles by adding keys already expanded to a stop list
	// 4. get a comma separted list as input, return a comma separated list as output
}
