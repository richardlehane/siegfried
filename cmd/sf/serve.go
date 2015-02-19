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

/*
simple HTTP protocol a la webdis. Single /identify route. Can do GET /identify/base64filename or POST /identify with file=file.
Params ?recurse=false&format=csv|yaml|json
/version
/identify/NAME?nr& GET
/identify/NAME?nr PUT file=body


import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/richardlehane/siegfried"
)

func handleError(w http.ResponseWriter, status int, e error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, e.Error())
}

func decodePath(s string) (string, error) {
	if len(s) < 11 {
		return "", fmt.Errorf("Path too short, expecting 11 characters got %d", len(s))
	}
	data, err := base64.URLEncoding.DecodeString(s[10:])
	if err != nil {
		return "", fmt.Errorf("Error base64 decoding file path, error message %v", err)
	}
	return string(data), nil
}

func identify(s *siegfried.Siegfried) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			rdr  io.ReadCloser
			path string
			err  error
		)
		if r.Method == "POST" {
			f, h, err := r.FormFile("file")
			if err != nil {
				return handleErr(err)
			}
			rdr = io.ReadCloser(f)
			path = h.Filename

		} else {
			path, err = decodePath(s.URL.Path)
			if err != nil {
				return handleErr(err)
			}
			f, err := os.Open(string(data))
			if err != nil {
				return handleErr(err)
			}
			rdr = io.ReadCloser(f)
		}
		defer rdr.Close()
		c, err := s.Identify(path, rdr)
		if err != nil {
			return handleErr(err)
		}
		for i := range c {
			io.WriteString(w, i.Json())
		}
	}
}

func handleMain(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" || r.URL.Path != "/" {
		return serve404(w)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return T("index").Execute(w, nil)
}

func displayVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "it works")
}

func server(port string, s *siegfried.Siegfried) {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/identify/", identify(s))
	http.ListenAndServe(port, nil)
}
*/
