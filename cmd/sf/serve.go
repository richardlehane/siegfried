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
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
)

func handleErr(w http.ResponseWriter, status int, e error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, fmt.Sprintf("SF server error; got %v\n", e))
}

func decodePath(s string) (string, error) {
	if len(s) < 11 {
		return "", fmt.Errorf("path too short, expecting 11 characters got %d", len(s))
	}
	return s[10:], nil
}

func parseRequest(w http.ResponseWriter, r *http.Request, s *siegfried.Siegfried, wg *sync.WaitGroup) (error, string, writer, bool, string, *siegfried.Siegfried, getFn) {
	// json, csv, droid or yaml
	paramsErr := func(field, expect string) (error, string, writer, bool, string, *siegfried.Siegfried, getFn) {
		return fmt.Errorf("bad request; in param %s got %s; valid values %s", field, r.FormValue(field), expect), "", nil, false, "", nil, nil
	}
	var (
		mime string
		wr   writer
		frmt int
	)
	switch {
	case *jsono:
		frmt = 1
	case *csvo:
		frmt = 2
	case *droido:
		frmt = 3
	}
	if v := r.FormValue("format"); v != "" {
		switch v {
		case "yaml":
			frmt = 0
		case "json":
			frmt = 1
		case "csv":
			frmt = 2
		case "droid":
			frmt = 3
		default:
			return paramsErr("format", "yaml, json, csv or droid")
		}
	}
	if accept := r.Header.Get("Accept"); accept != "" {
		switch accept {
		case "application/x-yaml":
			frmt = 0
		case "application/json":
			frmt = 1
		case "text/csv", "application/csv":
			frmt = 2
		case "application/x-droid":
			frmt = 3
		}
	}
	switch frmt {
	case 0:
		wr = newYAML(w)
		mime = "application/x-yaml"
	case 1:
		wr = newJSON(w)
		mime = "application/json"
	case 2:
		wr = newCSV(w)
		mime = "text/csv"
	case 3:
		wr = newDroid(w)
		mime = "application/x-droid"
	}
	// no recurse
	norec := *nr
	if v := r.FormValue("nr"); v != "" {
		switch v {
		case "true":
			norec = true
		case "false":
			norec = false
		default:
			paramsErr("nr", "true or false")
		}
	}
	// archive
	z := *archive
	if v := r.FormValue("z"); v != "" {
		switch v {
		case "true":
			z = true
		case "false":
			z = false
		default:
			paramsErr("z", "true or false")
		}
	}
	// checksum
	h := *hashf
	if v := r.FormValue("hash"); v != "" {
		h = v
	}
	cs := getHash(h)
	// sig
	sf := s
	if v := r.FormValue("sig"); v != "" {
		if _, err := os.Stat(config.Local(v)); err != nil {
			return fmt.Errorf("bad request; sig param should be path to a signature file (absolute or relative to home); got %v", err), "", nil, false, "", nil, nil
		}
		nsf, err := siegfried.Load(config.Local(v))
		if err == nil {
			sf = nsf
		}
	}
	gf := func(path, mime, mod string, sz int64) *context {
		c := ctxPool.Get().(*context)
		c.path, c.mime, c.mod, c.sz = path, mime, mod, sz
		c.s, c.w, c.wg, c.h, c.z = sf, wr, wg, cs, z
		return c
	}
	return nil, mime, wr, norec, h, sf, gf
}

func handleIdentify(s *siegfried.Siegfried, ctxts chan *context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		wg := &sync.WaitGroup{}
		err, mime, wr, nr, hh, sf, gf := parseRequest(w, r, s, wg)
		if err != nil {
			handleErr(w, http.StatusNotFound, err)
			return
		}
		if r.Method == "POST" {
			f, h, err := r.FormFile("file")
			if err != nil {
				handleErr(w, http.StatusNotFound, err)
				return
			}
			defer f.Close()
			var sz int64
			var mod string
			osf, ok := f.(*os.File)
			if ok {
				info, err := osf.Stat()
				if err != nil {
					handleErr(w, http.StatusInternalServerError, err)
				}
				sz = info.Size()
				mod = info.ModTime().String()
			} else {
				sz = r.ContentLength
			}
			w.Header().Set("Content-Type", mime)
			wr.writeHead(sf, hh)
			wg.Add(1)
			ctx := gf(h.Filename, "", mod, sz)
			ctxts <- ctx
			identifyRdr(f, ctx, ctxts, gf)
			wg.Wait()
			wr.writeTail()
			return
		} else {
			path, err := decodePath(r.URL.Path)
			if err == nil {
				_, err = os.Stat(path)
			}
			if err != nil {
				handleErr(w, http.StatusNotFound, err)
				return
			}
			w.Header().Set("Content-Type", mime)
			wr.writeHead(sf, hh)
			err = identify(ctxts, path, "", nr, gf)
			wg.Wait()
			wr.writeTail()
			if err != nil {
				if _, ok := err.(WalkError); ok { // only dump out walk errors, other errors reported in result
					io.WriteString(w, err.Error())
				}
			}
			return
		}
	}
}

const usage = `
	<html>
		<head>
			<title>Siegfried server</title>
		</head>
		<body>
			<h1><a name="top">Siegfried server usage</a></h1>
			<p>The siegfried server has two modes of identification:
			<ul><li><a href="#get_request">GET request</a>, where a file or directory path is given in the URL and the server retrieves the file(s);</li>
			<li><a href="#post_request">POST request</a>, where the file is sent over the network as form-data.</li></ul></p> 
			<h2>Default settings</h2>
			<p>When starting the server, you can use regular sf flags to set defaults for the <i>nr</i>, <i>format</i>, <i>hash</i>, <i>z</i>, and <i>sig</i> parameters that will apply to all requests unless overriden. Logging options can also be set.<p>
			<p>E.g. sf -nr -z -hash md5 -sig pronom-tika.sig -log warn,error -serve localhost:5138</p>
			<hr>
			<h2><a name="get_request">GET request</a></h2>
			<p><strong>GET</strong> <i>/identify/[file name or folder name (percent encoded)](?nr=true&format=yaml&hash=md5&z=true&sig=locfdd.sig)</i></p>
			<p>E.g. http://localhost:5138/identify/c%3A%2FUsers%2Frichardl%2FMy%20Documents%2Fhello%20world.docx?format=json</p>
			<h3>Parameters</h3>
			<p><i>nr</i> (optional) - stop sub-directory recursion when a directory path is given.</p>
			<p><i>format</i> (optional) - select the output format (csv, yaml, json, droid). Default is yaml. Alternatively, HTTP content negotiation can be used.</p>
			<p><i>hash</i> (optional) - calculate file checksum (md5, sha1, sha256, sha512, crc)</p>
			<p><i>z</i> (optional) - scan archive formats (zip, tar, gzip, warc, arc) with z=true. Default is false.</p>
			<p><i>sig</i> (optional) - load a specific signature file. Default is default.sig.</p>
			<h3>Example</h2>
			<!-- set the get target for the example form using js function at bottom page-->
			<h4>File/ directory:</h4>
			<p><input type="text" id="filename"> (provide the path to a file or directory e.g. c:\My Documents\file.doc. It will be percent encoded by this form.)</p>
			<h4>Parameters:</h4>
			<form method="get" id="get_example">
			 <p>No directory recursion (nr): <input type="radio" name="nr" value="true"> true <input type="radio" name="nr" value="false" checked> false</p>
			 <p>Format (format): <select name="format">
  				<option value="yaml">yaml</option>
  				<option value="json">json</option>
  				<option value="csv">csv</option>
 				<option value="droid">droid</option>
			</select></p>
			 <p>Hash (hash): <select name="hash">
  				<option value="none">none</option>
  				<option value="md5">md5</option>
  				<option value="sha1">sha1</option>
 				<option value="sha256">sha256</option>
 				<option value="sha512">sha512</option>
 				<option value="crc">crc</option>
			</select></p>
			 <p>Scan archive (z): <input type="radio" name="z" value="true"> true <input type="radio" name="z" value="false" checked> false</p>
			 <p>Signature file (sig): <input type="text" name="sig"></p>
			 <p><input type="submit" value="Submit"></p>
			</form>
			<p><a href="#top">Back to top</p>
			<hr>
			<h2><a name="post_request">POST request</a></h2>
			<p><strong>POST</strong> <i>/identify(?format=yaml&hash=md5&z=true&sig=locfdd.sig)</i> Attach a file as form-data with the key "file".</p>
			<p>E.g. curl "http://localhost:5138/identify?format=json&hash=crc" -F file=@myfile.doc</p>
			<h3>Parameters</h3>
			<p><i>format</i> (optional) - select the output format (csv, yaml, json, droid). Default is yaml. Alternatively, HTTP content negotiation can be used.</p>
			<p><i>hash</i> (optional) - calculate file checksum (md5, sha1, sha256, sha512, crc)</p>
			<p><i>z</i> (optional) - scan archive formats (zip, tar, gzip, warc, arc) with z=true. Default is false.</p>
			<p><i>sig</i> (optional) - load a specific signature file. Default is default.sig.</p>
			<h3>Example</h2>
			<form action="/identify" enctype="multipart/form-data" method="post">
			 <h4>File:</h4>
			 <p><input type="file" name="file"></p>
			 <h4>Parameters:</h4>
			 <p>Format (format): <select name="format">
  				<option value="yaml">yaml</option>
  				<option value="json">json</option>
  				<option value="csv">csv</option>
 				<option value="droid">droid</option>
			</select></p>
			 <p>Hash (hash): <select name="hash">
  				<option value="none">none</option>
  				<option value="md5">md5</option>
  				<option value="sha1">sha1</option>
 				<option value="sha256">sha256</option>
 				<option value="sha512">sha512</option>
 				<option value="crc">crc</option>
			</select></p>
			 <p>Scan archive (z): <input type="radio" name="z" value="true"> true <input type="radio" name="z" value="false" checked> false</p>
			 <p>Signature file (sig): <input type="text" name="sig"></p>
			 <p><input type="submit" value="Submit"></p>
			</form>
			<p><a href="#top">Back to top</p>
			<script>
				var input = document.getElementById('filename');
				input.addEventListener('input', function()
				{
					var frm = document.getElementById('get_example');
   				    frm.action = "/identify/" + encodeURIComponent(input.value);
				});
			</script>
		</body>
	</html>
`

func handleMain(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || r.URL.Path != "/" {
		handleErr(w, http.StatusNotFound, fmt.Errorf("Not a valid path"))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, usage)
}

func listen(port string, s *siegfried.Siegfried, ctxts chan *context) {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/identify", handleIdentify(s, ctxts))
	http.HandleFunc("/identify/", handleIdentify(s, ctxts))
	http.ListenAndServe(port, nil)
}
