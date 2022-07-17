// Copyright 2022 Richard Lehane. All rights reserved.
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

import "C"

import (
	"os"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/static"
)

func init() {
	sf = static.New()
}

var sf *siegfried.Siegfried

//export Identify
func Identify(path string) ([][][2]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	ids, err := sf.Identify(f, path, "")
	if err != nil {
		return nil, err
	}
	ret := make([][][2]string, len(ids))
	for i := range ret {
		ret[i] = sf.Label(ids[i])
	}
	return ret, nil
}

func main() {}
