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
	"time"

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
		//return newJSON(f, path)
	case 'O', 'K':
		//return newFido(f, path)
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
