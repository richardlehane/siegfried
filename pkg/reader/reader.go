// Copyright 2017 Richard Lehane. All rights reserved.
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

package reader

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/internal/checksum"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

const (
	unknownWarn = "no match"
	extWarn     = "match on extension only"
	extMismatch = "extension mismatch"
)

type Reader interface {
	Head() Head
	Next() (File, error)
}

type Head struct {
	ResultsPath   string
	SignaturePath string
	Scanned       time.Time
	Created       time.Time
	Version       [3]int
	Identifiers   [][2]string
	Fields        [][]string
	HashHeader    string
}

type File struct {
	Path string
	Size int64
	Mod  time.Time
	Hash []byte
	Err  error
	IDs  []core.Identification
}

type record struct {
	attributes map[string]string
	listFields []string
	listValues []string
}

func toVersion(str string) ([3]int, error) {
	var ret [3]int
	if str == "" {
		return ret, nil
	}
	nums := strings.Split(str, ".")
	if len(nums) != len(ret) {
		return ret, fmt.Errorf("bad version; got %d numbers", len(nums))
	}
	for i, v := range nums {
		var err error
		ret[i], err = strconv.Atoi(v)
		if err != nil {
			return ret, fmt.Errorf("bad version; got %v", err)
		}
	}
	return ret, nil
}

func getHead(rec record) (Head, error) {
	head, err := newHeadMap(rec.attributes)
	head.Identifiers = getIdentifiers(rec.listValues)
	return head, err
}

func newHeadMap(m map[string]string) (Head, error) {
	return newHead(m["results"], m["signature"], m["scandate"], m["created"], m["siegfried"])
}

func newHead(resultsPath, sigPath, scanned, created, version string) (Head, error) {
	var err error
	h := Head{
		ResultsPath:   resultsPath,
		SignaturePath: sigPath,
	}
	h.Version, err = toVersion(version)
	if scanned != "" {
		h.Scanned, err = time.Parse(time.RFC3339, scanned)
	}
	if created != "" {
		h.Created, err = time.Parse(time.RFC3339, created)
	}
	return h, err
}

func newFile(path, sz, mod, hash, e string) (File, error) {
	var err error
	file := File{
		Path: path,
		IDs:  make([]core.Identification, 0, 1),
	}
	if mod != "" {
		file.Mod, err = time.Parse(time.RFC3339, mod)
		if err != nil {
			file.Mod, err = time.Parse(droidTime, mod)
		}
	}
	if err != nil {
		err = fmt.Errorf("bad field, mod: %s, err: %v", mod, err)
	}
	if len(hash) > 0 {
		file.Hash = []byte(hash)
	}
	if e != "" {
		file.Err = fmt.Errorf("%s", e)
	}
	if sz == "" {
		return file, nil
	}
	fs, fserr := strconv.Atoi(sz)
	file.Size = int64(fs)
	if fserr != nil {
		err = fmt.Errorf("bad field, sz: %s, err: %v", sz, err)
	}
	return file, err
}

func getFile(rec record) (File, error) {
	var hh string
	for k := range rec.attributes {
		if h := checksum.GetHash(k); h >= 0 {
			hh = k
			break
		}
	}
	f, err := newFile(rec.attributes["filename"],
		rec.attributes["filesize"],
		rec.attributes["modified"],
		rec.attributes[hh],
		rec.attributes["errors"],
	)
	if err != nil {
		return f, err
	}
	var sidx, eidx int
	for i, v := range rec.listFields {
		if v == "ns" {
			eidx = i
			if eidx > sidx {
				f.IDs = append(f.IDs, newDefaultID(rec.listFields[sidx:eidx], rec.listValues[sidx:eidx]))
				sidx = eidx
			}
		}
	}
	f.IDs = append(f.IDs, newDefaultID(rec.listFields[sidx:len(rec.listFields)], rec.listValues[sidx:len(rec.listFields)]))
	return f, nil
}

func getIdentifiers(vals []string) [][2]string {
	ret := make([][2]string, 0, len(vals)/2)
	for i, v := range vals {
		if i%2 == 0 {
			ret = append(ret, [2]string{v, ""})
		} else {
			ret[len(ret)-1][1] = v
		}
	}
	return ret
}

func getHash(m map[string]string) string {
	for k := range m {
		if h := checksum.GetHash(k); h >= 0 {
			return h.String()
		}
	}
	return ""
}

func getFields(keys, vals []string) [][]string {
	ret := make([][]string, 0, 1)
	var ns string
	var consume bool
	for i, v := range keys {
		if v == "ns" || v == "namespace" {
			if ns == vals[i] {
				consume = false
			} else {
				ns = vals[i]
				consume = true
				ret = append(ret, []string{})
				v = "namespace" // always store as namespace
			}
		}
		if consume {
			ret[len(ret)-1] = append(ret[len(ret)-1], v)
		}
	}
	return ret
}

type peekReader struct {
	unread bool
	peek   byte
	rdr    io.Reader
}

func (pr *peekReader) Read(b []byte) (int, error) {
	if pr.unread {
		if len(b) < 1 {
			return 0, nil
		}
		b[0] = pr.peek
		pr.unread = false
		if len(b) == 1 {
			return 1, nil
		}
		i, e := pr.rdr.Read(b[1:])
		return i + 1, e
	}
	return pr.rdr.Read(b)
}

func New(rdr io.Reader, path string) (Reader, error) {
	buf := make([]byte, 1)
	if _, err := rdr.Read(buf); err != nil {
		return nil, err
	}
	pr := &peekReader{true, buf[0], rdr}
	switch buf[0] {
	case '-':
		return newYAML(pr, path)
	case 'f':
		return newCSV(pr, path)
	case '{':
		return newJSON(pr, path)
	case 'O', 'K':
		return newFido(pr, path)
	case 'D':
		return newDroidNp(pr, path)
	case '"', 'I':
		return newDroid(pr, path)
	}
	return nil, fmt.Errorf("not a valid results file, bad char %d", int(buf[0]))
}

type defaultID struct {
	id     int
	warn   int
	known  bool
	values []string
}

func (did *defaultID) String() string { return did.values[did.id] }
func (did *defaultID) Known() bool    { return did.known }
func (did *defaultID) Warn() string {
	if did.warn > 0 {
		return did.values[did.warn]
	}
	return ""
}
func (did *defaultID) Values() []string        { return did.values }
func (did *defaultID) Archive() config.Archive { return config.None }

func newDefaultID(fields, values []string) *defaultID {
	did := &defaultID{values: values}
	for i, v := range fields {
		switch v {
		case "id", "identifier", "ID":
			did.id = i
			switch values[i] {
			case "unknown", "UNKNOWN", "":
			default:
				did.known = true
			}
		case "warn", "warning":
			did.warn = i
		}
	}
	return did
}
