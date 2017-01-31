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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/cmd/internal/checksum"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

type Reader interface {
	Head() Head
	Next() (File, error)
	Close() error
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
	Mod  string
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
	file := File{
		Path: path,
		Mod:  mod,
		IDs:  make([]core.Identification, 0, 1),
	}
	if len(hash) > 0 {
		file.Hash = []byte(hash)
	}
	if e != "" {
		file.Err = fmt.Errorf("%s", e)
	}
	size, err := strconv.Atoi(sz)
	if err != nil {
		return file, fmt.Errorf("bad field; expecting int got %s", sz)
	}
	file.Size = int64(size)
	return file, nil
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
		if v == "ns" {
			if ns == vals[i] {
				consume = false
			} else {
				ns = vals[i]
				consume = true
				ret = append(ret, []string{})
			}
		}
		if consume {
			ret[len(ret)-1] = append(ret[len(ret)-1], v)
		}
	}
	return ret
}

func Open(path string) (Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 1)
	f.ReadAt(buf, 0)
	switch buf[0] {
	case '-':
		return newYAML(f, path)
	case 'f':
		return newCSV(f, path)
	case '{':
		return newJSON(f, path)
	case 'O', 'K':
		return newFido(f, path)
	case 'D':
		//return newDroidNp(f, path)
	case '"':
		//return newDroid(f, path)
	}
	return nil, fmt.Errorf("not a valid results file")
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
