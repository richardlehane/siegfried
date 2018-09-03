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

var (
	// list of flags that can be configured
	setableFlags = []string{"coe", "csv", "droid", "hash", "json", "log", "multi", "nr", "serve", "sig", "throttle", "yaml", "z"}
	// list of flags that control output - these are exclusive of each other
	outputFlags = []string{"csv", "droid", "json", "yaml"}
)

// also used in sf_test.go
func check(s string, ss []string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// if -setconf flag set, write settable flags to a conf file. Returns flag names set and an error.
func setconf() (string, error) {
	buf := &bytes.Buffer{}
	var settables []string
	flag.Visit(func(fl *flag.Flag) {
		if !check(fl.Name, setableFlags) {
			return
		}
		fmt.Fprintf(buf, "%s:%s\n", fl.Name, fl.Value.String())
		settables = append(settables, fl.Name)
	})
	if len(settables) > 0 {
		return strings.Join(settables, ", "), ioutil.WriteFile(config.Conf(), buf.Bytes(), 0644)
	}
	// no flags - so we delete the conf file if it exists
	if _, err := os.Stat(config.Conf()); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return "", os.Remove(config.Conf())
}

// if it exists, read defaults from the conf file.
func getconf() (map[string]string, error) {
	if _, err := os.Stat(config.Conf()); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	f, err := os.Open(config.Conf())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ret := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		kv := strings.SplitN(scanner.Text(), ":", 2)
		if len(kv) != 2 {
			continue
		}
		ret[kv[0]] = kv[1]
	}
	return ret, nil
}

// if it exists, read defaults from the conf file. Overwrite defaults with any flags explictly set
func readconf() error {
	confFlags, err := getconf()
	if len(confFlags) == 0 {
		return err
	}
	// remove conf values for any flags explictly set
	flag.Visit(func(fl *flag.Flag) {
		// if an output flag has been explicitly set, delete any that may be in the conf file
		if check(fl.Name, outputFlags) {
			for _, v := range outputFlags {
				delete(confFlags, v)
			}
		} else {
			delete(confFlags, fl.Name)
		}
	})
	for k, v := range confFlags {
		if err = flag.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}
