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
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

type iterableID interface {
	next() core.Identification
}

type idChan chan core.Identification

func (ids idChan) next() core.Identification {
	id, ok := <-ids
	if !ok {
		return nil
	}
	return id
}

type idSlice struct {
	idx int
	ids []core.Identification
}

func (is *idSlice) next() core.Identification {
	is.idx++
	if is.idx > len(is.ids) {
		return nil
	}
	return is.ids[is.idx-1]
}

type writer interface {
	writeHead(s *siegfried.Siegfried)
	// if a directory give a negative sz
	writeFile(name string, sz int64, mod string, checksum []byte, err error, ids iterableID) config.Archive
	writeTail()
}

type debugWriter struct{}

func (d debugWriter) writeHead(s *siegfried.Siegfried) {}
func (d debugWriter) writeFile(name string, sz int64, mod string, cs []byte, err error, ids iterableID) config.Archive {
	return 0
}
func (d debugWriter) writeTail() {}

type csvWriter struct {
	rec []string
	w   *csv.Writer
}

func newCSV(w io.Writer) *csvWriter {
	return &csvWriter{make([]string, 11), csv.NewWriter(os.Stdout)}
}

func (c *csvWriter) writeHead(s *siegfried.Siegfried) {
	c.w.Write([]string{"filename", "filesize", "file modified", "errors", "identifier", "id", "format name", "format version", "mimetype", "basis", "warning"})
}

func (c *csvWriter) writeFile(name string, sz int64, mod string, checksum []byte, err error, ids iterableID) config.Archive {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	if ids == nil {
		empty := make([]string, 7)
		c.rec[0], c.rec[1], c.rec[2], c.rec[3] = name, strconv.Itoa(int(sz)), mod, errStr
		copy(c.rec[4:], empty)
		c.w.Write(c.rec)
		return 0
	}
	var archive config.Archive
	for id := ids.next(); id != nil; id = ids.next() {
		if id.Archive() > archive {
			archive = id.Archive()
		}
		c.rec[0], c.rec[1], c.rec[2], c.rec[3] = name, strconv.Itoa(int(sz)), mod, errStr
		copy(c.rec[4:], id.CSV())
		c.w.Write(c.rec)
	}
	return archive
}

func (c *csvWriter) writeTail() { c.w.Flush() }

type yamlWriter struct {
	replacer *strings.Replacer
	w        *bufio.Writer
}

func newYAML(w io.Writer) *yamlWriter {
	return &yamlWriter{strings.NewReplacer("'", "''"), bufio.NewWriter(w)}
}

func (y *yamlWriter) writeHead(s *siegfried.Siegfried) {
	y.w.WriteString(s.YAML())
}

func (y *yamlWriter) writeFile(name string, sz int64, mod string, checksum []byte, err error, ids iterableID) config.Archive {
	var errStr string
	if err != nil {
		errStr = fmt.Sprintf("'%s'", err.Error())
	}
	fmt.Fprintf(y.w, "---\nfilename : '%s'\nfilesize : %d\nmodified : %s\nerrors   : %s\nmatches  :\n", y.replacer.Replace(name), sz, mod, errStr)
	if ids == nil {
		return 0
	}
	var archive config.Archive
	for id := ids.next(); id != nil; id = ids.next() {
		if id.Archive() > archive {
			archive = id.Archive()
		}
		y.w.WriteString(id.YAML())
	}
	return archive
}

func (y *yamlWriter) writeTail() { y.w.Flush() }

type jsonWriter struct {
	subs     bool
	replacer *strings.Replacer
	w        *bufio.Writer
}

func newJSON(w io.Writer) *jsonWriter {
	return &jsonWriter{false, strings.NewReplacer(`"`, `\"`, `\\`, `\\`, `\`, `\\`), bufio.NewWriter(w)}
}

func (j *jsonWriter) writeHead(s *siegfried.Siegfried) {
	j.w.WriteString(s.JSON())
	j.w.WriteString("\"files\":[")
}

func (j *jsonWriter) writeFile(name string, sz int64, mod string, checksum []byte, err error, ids iterableID) config.Archive {
	if j.subs {
		j.w.WriteString(",")
	}
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	fmt.Fprintf(j.w, "{\"filename\":\"%s\",\"filesize\": %d,\"modified\":\"%s\",\"errors\": \"%s\",\"matches\": [", j.replacer.Replace(name), sz, mod, errStr)
	if ids == nil {
		return 0
	}
	var subs bool
	var archive config.Archive
	for id := ids.next(); id != nil; id = ids.next() {
		if id.Archive() > archive {
			archive = id.Archive()
		}
		if subs {
			j.w.WriteString(",")
		}
		j.w.WriteString(id.JSON())
		subs = true
	}
	j.w.WriteString("]}")
	j.subs = true
	return archive
}

func (j *jsonWriter) writeTail() {
	j.w.WriteString("]}\n")
	j.w.Flush()
}

type droidWriter struct {
	id      int
	parents map[string]int
	rec     []string
	w       *csv.Writer
}

func newDroid(w io.Writer) *droidWriter {
	return &droidWriter{
		parents: make(map[string]int),
		rec:     make([]string, 18),
		w:       csv.NewWriter(os.Stdout),
	}
}

// "identifier", "id", "format name", "format version", "mimetype", "basis", "warning"

func (d *droidWriter) writeHead(s *siegfried.Siegfried) {
	d.w.Write([]string{
		"ID", "PARENT_ID", "URI", "FILE_PATH", "NAME",
		"METHOD", "STATUS", "SIZE", "TYPE", "EXT",
		"LAST_MODIFIED", "EXTENSION_MISMATCH", hashHeader(), "FORMAT_COUNT",
		"PUID", "MIME_TYPE", "FORMAT_NAME", "FORMAT_VERSION"})
}

func (d *droidWriter) writeFile(name string, sz int64, mod string, checksum []byte, err error, ids iterableID) config.Archive {
	d.id++
	errStr := "Done"
	if err != nil {
		errStr = err.Error()
	}
	if ids == nil {
		empty := make([]string, 7)
		d.rec[0], d.rec[1], d.rec[2] = strconv.Itoa(d.id), strconv.Itoa(int(sz)), errStr
		copy(d.rec[3:], empty)
		d.w.Write(d.rec)
		return 0
	}
	var archive config.Archive
	for id := ids.next(); id != nil; id = ids.next() {
		if id.Archive() > archive {
			archive = id.Archive()
		}
		d.rec[0], d.rec[1], d.rec[2] = name, strconv.Itoa(int(sz)), errStr
		copy(d.rec[3:], id.CSV())
		d.w.Write(d.rec)
	}
	return archive
}

func (d *droidWriter) writeTail() { d.w.Flush() }

func processPath(path string) (uri, full, dir, base, ext string) {
	full, _ = filepath.Abs(path)
	dir = filepath.Dir(path)
	base = filepath.Base(path)
	ext = filepath.Ext(path)
	uri = "file:///" + filepath.ToSlash(full)
	return
}

func getMethod(basis string) string {
	switch {
	case strings.Contains(basis, "container"):
		return "Container"
	case strings.Contains(basis, "byte"):
		return "Signature"
	case strings.Contains(basis, "extension"):
		return "Extension"
	case strings.Contains(basis, "text"):
		return "Text"
	}
	return ""
}

func mismatch(warning string) string {
	if strings.Contains(warning, "extension mismatch") {
		return "TRUE"
	}
	return "FALSE"
}
