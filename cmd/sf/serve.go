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
*/

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/richardlehane/siegfried"
)

func handleErr(w http.ResponseWriter, status int, e error) {
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

func parseRequest(w http.ResponseWriter, r *http.Request) (string, writer, bool) {
	var nr bool
	vals := r.URL.Query()
	if n, ok := vals["nr"]; ok && len(n) > 0 {
		if n[0] == "true" {
			nr = true
		}
	}
	if format, ok := vals["format"]; ok && len(format) > 0 {
		switch format[0] {
		case "json":
			return "application/json", newJson(w), nr
		case "csv":
			return "text/csv", newCsv(w), nr
		case "yaml":
			return "application/x-yaml", newYaml(w), nr
		}
	}
	if accept := r.Header.Get("Accept"); accept != "" {
		switch accept {
		case "application/json":
			return "application/json", newJson(w), nr
		case "text/csv", "application/csv":
			return "text/csv", newCsv(w), nr
		case "application/x-yaml":
			return "application/x-yaml", newYaml(w), nr
		}
	}
	return "application/json", newJson(w), nr
}

func identify(s *siegfried.Siegfried) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mime, wr, nr := parseRequest(w, r)
		if r.Method == "POST" {
			f, h, err := r.FormFile("file")
			if err != nil {
				handleErr(w, http.StatusNotFound, err)
				return
			}
			defer f.Close()
			var sz int64
			osf, ok := f.(*os.File)
			if ok {
				info, err := osf.Stat()
				if err != nil {
					handleErr(w, http.StatusInternalServerError, err)
				}
				sz = info.Size()
			} else {
				sz = r.ContentLength
			}
			w.Header().Set("Content-Type", mime)
			wr.writeHead(s)
			c, err := s.Identify(h.Filename, f)
			if c == nil {
				wr.writeFile(h.Filename, sz, fmt.Errorf("failed to identify %s, got: %v", h.Filename, err), nil)
				return
			}
			wr.writeFile(h.Filename, sz, err, idChan(c))
			wr.writeTail()
			return
		} else {
			path, err := decodePath(r.URL.Path)
			if err != nil {
				handleErr(w, http.StatusNotFound, err)
				return
			}
			info, err := os.Stat(path)
			if err != nil {
				handleErr(w, http.StatusNotFound, err)
				return
			}
			w.Header().Set("Content-Type", mime)
			wr.writeHead(s)
			if info.IsDir() {
				multiIdentifyS(wr, s, path, nr)
				wr.writeTail()
				return
			}
			identifyFile(wr, s, path, info.Size())
			wr.writeTail()
			return
		}

	}
}

const welcome = `
	<html>
		<head>
			<title>Siegfried server</title>
		</head>
		<body>
			<h1>Siegfried server usage</h1>
			<p><strong>GET</strong> <i>/identify/[<a href="https://tools.ietf.org/html/rfc4648#section-5">URL-safe base64 encoded</a> file name or folder name](?nr=true&format=csv|yaml|json)</i></p>
			<p><strong>POST</strong> <i>/identify(?format=csv|yaml|json)</i></p>
			<p>If using POST, attach a file as form-data with the key "file". E.g. <i>curl localhost:8080/identify -F file=@myfile.doc</i></p>
		</body>
	</html>
`

func handleMain(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || r.URL.Path != "/" {
		handleErr(w, http.StatusNotFound, fmt.Errorf("Not a valid path"))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, welcome)
}

func listen(port string, s *siegfried.Siegfried) {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/identify", identify(s))
	http.HandleFunc("/identify/", identify(s))
	http.ListenAndServe(port, nil)
}
