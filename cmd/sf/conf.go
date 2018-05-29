// Copyright 2018 Richard Lehane. All rights reserved.
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
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/richardlehane/siegfried/pkg/config"
)

// list of flags that can be configured
var setableFlags = []string{"coe", "csv", "droid", "hash", "json", "log", "multi", "nr", "serve", "sig", "throttle", "z"}

func setable(f string) bool {
	for _, v := range setableFlags {
		if f == v {
			return true
		}
	}
	return false
}

func setconf() error {
	buf := &bytes.Buffer{}
	flag.Visit(func(fl *flag.Flag) {
		if !setable(fl.Name) {
			return
		}
		fmt.Fprintf(buf, "%s:%s\n", fl.Name, fl.Value.String())
	})
	return ioutil.WriteFile(config.Conf(), buf.Bytes(), 0644)
}

func readconf() error {
	if _, err := os.Stat(config.Conf()); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	f, err := os.Open(config.Conf())
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	confFlags := make(map[string]string)
	for scanner.Scan() {
		kv := strings.SplitN(scanner.Text(), ":", 2)
		if len(kv) != 2 {
			continue
		}
		confFlags[kv[0]] = kv[1]
	}
	// remove conf values for any flags explictly set
	flag.Visit(func(fl *flag.Flag) {
		delete(confFlags, fl.Name)
	})
	for k, v := range confFlags {
		if err = flag.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}
