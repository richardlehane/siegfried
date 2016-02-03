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

// Warcscan is a simple script to enable searching WARC files and retrieving
// individual WARC records. Built it to generate test files with interesting
// characteristics e.g. presence of continuations or Content-Encoding.
//
// This is a very basic implementation that just scans on "WARC/" to split WARC
// files. Because using bufio.Scanner, cannot retrieve very large WARC records
// in this way (60kb is defined as limit to prevent panics).
//
// Once built, use e.g. `warcscan -s termone,termtwo -a -o output.warc input.warc`
// The -s flag is a comma-separated list of search terms.
// The -a flag says all terms must be matched (otherwise is an OR search).
// The -o flag names the output file (defaults to out.warc)
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	search = flag.String("s", "", "enter comma-separated search terms")
	out    = flag.String("o", "out.warc", "enter name of output file")
	all    = flag.Bool("a", false, "all search terms must match")
)

var marker = []byte("WARC/")

const maxbuf = 60000

func main() {
	flag.Parse()

	searchStrings := bytes.Split([]byte(*search), []byte(","))
	if len(searchStrings) == 0 {
		fmt.Println("must provide comma-separated search terms for scanning, with -s flag")
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Println("must provide an input warc file for scanning")
		os.Exit(0)
	}

	f, e := os.Open(flag.Arg(0))
	if e != nil {
		fmt.Println(e)
		os.Exit(0)
	}

	outBuf := &bytes.Buffer{}
	mainBuf := bufio.NewScanner(f)
	var overran bool

	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, marker); i >= 0 {
			overran = false
			return i + len(marker), data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		if len(data) > maxbuf {
			overran = true
			return len(data), data, nil
		}
		return 0, nil, nil
	}

	mainBuf.Split(split)
	for mainBuf.Scan() {
		if overran {
			continue
		}
		for idx, v := range searchStrings {
			if i := bytes.Index(mainBuf.Bytes(), v); i >= 0 {
				if *all && idx < len(searchStrings)-1 {
					continue
				}
				_, err := outBuf.Write(marker)
				if err == nil {
					_, err = outBuf.Write(mainBuf.Bytes())
				}
				if err != nil {
					fmt.Println(err)
					os.Exit(0)
				}
				break
			}
			if *all {
				break
			}
		}
	}

	ioutil.WriteFile(*out, outBuf.Bytes(), 0666)
}
