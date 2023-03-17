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
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/internal/checksum"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/writer"
)

func handleErr(w http.ResponseWriter, status int, e error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, fmt.Sprintf("SF server error; got %v\n", e))
}

func decodePath(s, b64 string) (string, error) {
	if len(s) < 11 {
		return "", fmt.Errorf("path too short, expecting at least 11 characters got %d", len(s))
	}
	if b64 == "true" {
		data, err := base64.URLEncoding.DecodeString(s[10:])
		if err != nil {
			return "", fmt.Errorf("Error base64 decoding file path, error message %v", err)
		}
		return string(data), nil
	}
	return s[10:], nil
}

func parseRequest(w http.ResponseWriter, r *http.Request, s *siegfried.Siegfried, wg *sync.WaitGroup) (string, writer.Writer, bool, bool, bool, checksum.HashTyp, *siegfried.Siegfried, getFn, error) {
	// json, csv, droid or yaml
	paramsErr := func(field, expect string) (string, writer.Writer, bool, bool, bool, checksum.HashTyp, *siegfried.Siegfried, getFn, error) {
		return "", nil, false, false, false, -1, nil, nil, fmt.Errorf("bad request; in param %s got %s; valid values %s", field, r.FormValue(field), expect)
	}
	var (
		mime string
		wr   writer.Writer
		d    bool
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
		wr = writer.YAML(w)
		mime = "application/x-yaml"
	case 1:
		wr = writer.JSON(w)
		mime = "application/json"
	case 2:
		wr = writer.CSV(w)
		mime = "text/csv"
	case 3:
		wr = writer.Droid(w)
		d = true
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
	// continue on error
	coerr := *coe
	if v := r.FormValue("coe"); v != "" {
		switch v {
		case "true":
			coerr = true
		case "false":
			coerr = false
		default:
			paramsErr("coe", "true or false")
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
	ht := checksum.GetHash(h)
	// sig
	sf := s
	if v := r.FormValue("sig"); v != "" {
		if _, err := os.Stat(config.Local(v)); err != nil {
			return "", nil, false, false, false, -1, nil, nil, fmt.Errorf("bad request; sig param should be path to a signature file (absolute or relative to home); got %v", err)
		}
		nsf, err := siegfried.Load(config.Local(v))
		if err == nil {
			sf = nsf
		}
	}
	gf := func(path, mime string, mod time.Time, sz int64) *context {
		c := ctxPool.Get().(*context)
		c.path, c.mime, c.mod, c.sz = path, mime, mod, sz
		c.s, c.wg, c.w, c.d, c.z, c.h = sf, wg, wr, d, z, checksum.MakeHash(ht)
		return c
	}
	return mime, wr, coerr, norec, d, ht, sf, gf, nil
}

func handleIdentify(w http.ResponseWriter, r *http.Request, s *siegfried.Siegfried, ctxts chan *context) {
	wg := &sync.WaitGroup{}
	mime, wr, coerr, nrec, d, ht, sf, gf, err := parseRequest(w, r, s, wg)
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
		var mod time.Time
		osf, ok := f.(*os.File)
		if ok {
			info, err := osf.Stat()
			if err != nil {
				handleErr(w, http.StatusInternalServerError, err)
			}
			sz = info.Size()
			mod = info.ModTime()
		} else {
			sz = r.ContentLength
		}
		w.Header().Set("Content-Type", mime)
		wr.Head(config.SignatureBase(), time.Now(), sf.C, config.Version(), sf.Identifiers(), sf.Fields(), ht.String())
		wg.Add(1)
		ctx := gf(h.Filename, "", mod, sz)
		ctxts <- ctx
		identifyRdr(f, ctx, ctxts, gf)
		wg.Wait()
		wr.Tail()
		return
	}
	path, err := decodePath(r.URL.Path, r.FormValue("base64"))
	if err == nil {
		_, err = os.Stat(path)
	}
	if err != nil {
		handleErr(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", mime)
	wr.Head(config.SignatureBase(), time.Now(), sf.C, config.Version(), sf.Identifiers(), sf.Fields(), ht.String())
	err = identify(ctxts, path, "", coerr, nrec, d, gf)
	wg.Wait()
	wr.Tail()
	if _, ok := err.(WalkError); ok { // only dump out walk errors, other errors reported in result
		io.WriteString(w, err.Error())
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
			<p>The update command can also be issued as a GET request to <a href="/update">/update</a>. This fetches an updated signature file and hot patches the running siegfried instance.</p>
			<p>If PRONOM isn't being used as the underlying identifier, the update command can be qualified with the name of a different identifer e.g. <a href="/update">/update/wikidata</a>.</p>
			<h2>Default settings</h2>
			<p>When starting the server, you can use regular sf flags to set defaults for the <i>nr</i>, <i>format</i>, <i>hash</i>, <i>z</i>, and <i>sig</i> parameters that will apply to all requests unless overridden. Logging options can also be set.<p>
			<p>E.g. sf -nr -z -hash md5 -sig pronom-tika.sig -log p,w,e -serve localhost:5138</p>
			<hr>
			<h2><a name="get_request">GET request</a></h2>
			<p><strong>GET</strong> <i>/identify/[file or folder name (percent encoded)](?base64=false&nr=true&format=yaml&hash=md5&z=true&sig=locfdd.sig)</i></p>
			<p>E.g. http://localhost:5138/identify/c%3A%2FUsers%2Frichardl%2FMy%20Documents%2Fhello%20world.docx?format=json</p>
			<h3>Parameters</h3>
			<p><i>base64</i> (optional) - use <a href="https://tools.ietf.org/html/rfc4648#section-5">URL-safe base64 encoding</a> for the file or folder name with base64=true.</p>
			<p><i>coe</i> (optional) - continue directory scans even when fatal file access errors are encountered with coe=true.</p>
			<p><i>nr</i> (optional) - stop sub-directory recursion when a directory path is given with nr=true.</p>
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
			  <p>Use base64 encoding (base64): <input type="radio" name="base64" value="true"> true <input type="radio" name="base64" value="false" checked> false</p>
			 <p>Continue on error (coe): <input type="radio" name="coe" value="true"> true <input type="radio" name="nr" value="false" checked> false</p>
			 <p>No directory recursion (nr): <input type="radio" name="nr" value="true"> true <input type="radio" name="nr" value="false" checked> false</p>
			 <p>Format (format): <select name="format">
  				<option value="json">json</option>
  				<option value="yaml">yaml</option>
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
  				<option value="json">json</option>
  				<option value="yaml">yaml</option>
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
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, usage)
}

func handleUpdate(w http.ResponseWriter, r *http.Request, m *muxer) {
	args := []string{}
	if len(r.URL.Path) > 8 {
		args = append(args, r.URL.Path[8:])
	}
	updated, msg, err := updateSigs("", args)
	if err != nil {
		handleErr(w, http.StatusInternalServerError, err)
		return
	}
	if updated {
		defer func() {
			if r := recover(); r != nil {
				handleErr(w, http.StatusInternalServerError, fmt.Errorf("panic: %v", r))
			}
		}()
		nsf, err := siegfried.Load(config.Signature()) // may panic
		if err == nil {
			m.s = nsf // hot swap the siegfried!
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			io.WriteString(w, msg)
			return
		} else {
			handleErr(w, http.StatusInternalServerError, err)
			return
		}
	}
	w.WriteHeader(http.StatusNotModified)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, msg)
}

type muxer struct {
	s     *siegfried.Siegfried
	ctxts chan *context
	mut   sync.RWMutex
}

func (m *muxer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if (len(r.URL.Path) == 0 || r.URL.Path == "/") && r.Method == "GET" {
		handleMain(w, r)
		return
	}
	if len(r.URL.Path) >= 9 && r.URL.Path[:9] == "/identify" {
		m.mut.RLock()
		handleIdentify(w, r, m.s, m.ctxts)
		m.mut.RUnlock()
		return
	}
	if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/update" {
		m.mut.Lock()
		handleUpdate(w, r, m)
		m.mut.Unlock()
		return
	}
	handleErr(w, http.StatusNotFound, fmt.Errorf("valid paths are /, /update, /update/*, /identify and /identify/*"))
}

func listen(port string, s *siegfried.Siegfried, ctxts chan *context) {
	mux := &muxer{
		s:     s,
		ctxts: ctxts,
	}
	http.ListenAndServe(port, mux)
}
