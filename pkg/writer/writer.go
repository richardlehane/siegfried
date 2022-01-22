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

package writer

import (
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

type Writer interface {
	Head(path string, scanned, created time.Time, version [3]int, ids [][2]string, fields [][]string, hh string) // 	path := filepath.Base(path)
	File(name string, sz int64, mod string, checksum []byte, err error, ids []core.Identification)               // if a directory give a negative sz
	Tail()
}

func Null() Writer {
	return null{}
}

type null struct{}

func (n null) Head(path string, scanned, created time.Time, version [3]int, ids [][2]string, fields [][]string, hh string) {
}
func (n null) File(name string, sz int64, mod string, cs []byte, err error, ids []core.Identification) {
}
func (n null) Tail() {}

type csvWriter struct {
	recs  [][]string
	names []string
	w     *csv.Writer
}

func CSV(w io.Writer) Writer {
	return &csvWriter{w: csv.NewWriter(w)}
}

func (c *csvWriter) Head(path string, scanned, created time.Time, version [3]int, ids [][2]string, fields [][]string, hh string) {
	c.names = make([]string, len(fields))
	l := 4
	if hh != "" {
		l++
	}
	for i, f := range fields {
		l += len(f)
		c.names[i] = f[0]
	}
	c.recs = make([][]string, 1)
	c.recs[0] = make([]string, l)
	c.recs[0][0], c.recs[0][1], c.recs[0][2], c.recs[0][3] = "filename", "filesize", "modified", "errors"
	idx := 4
	if hh != "" {
		c.recs[0][4] = hh
		idx++
	}
	for _, f := range fields {
		copy(c.recs[0][idx:], f)
		idx += len(f)
	}
	c.w.Write(c.recs[0])
}

func (c *csvWriter) File(name string, sz int64, mod string, checksum []byte, err error, ids []core.Identification) {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	c.recs[0][0], c.recs[0][1], c.recs[0][2], c.recs[0][3] = name, strconv.FormatInt(sz, 10), mod, errStr
	idx := 4
	if checksum != nil {
		c.recs[0][4] = hex.EncodeToString(checksum)
		idx++
	}
	if len(ids) == 0 {
		empty := make([]string, len(c.recs[0])-idx)
		if checksum != nil {
			c.recs[0][4] = ""
		}
		copy(c.recs[0][idx:], empty)
		c.w.Write(c.recs[0])
		return
	}

	var thisName string
	var rowIdx, colIdx, prevLen int
	colIdx = idx
	for _, id := range ids {
		fields := id.Values()
		if thisName == fields[0] {
			rowIdx++
		} else {
			thisName = fields[0]
			rowIdx = 0
			colIdx += prevLen
			prevLen = len(fields)
		}
		if rowIdx >= len(c.recs) {
			c.recs = append(c.recs, make([]string, len(c.recs[0])))
			copy(c.recs[rowIdx][:idx], c.recs[0][:idx])
		}
		copy(c.recs[rowIdx][colIdx:], fields)
	}
	for _, r := range c.recs {
		c.w.Write(r)
	}
	c.recs = c.recs[:1]
}

func (c *csvWriter) Tail() { c.w.Flush() }

type yamlWriter struct {
	replacer *strings.Replacer
	w        *bufio.Writer
	hh       string
	hstrs    []string
	vals     [][]interface{}
}

func YAML(w io.Writer) Writer {
	return &yamlWriter{
		replacer: strings.NewReplacer("'", "''"),
		w:        bufio.NewWriter(w),
	}
}

func header(fields []string) string {
	headings := make([]string, len(fields))
	var max int
	for _, v := range fields {
		if v != "namespace" && len(v) > max {
			max = len(v)
		}
	}
	pad := fmt.Sprintf("%%-%ds", max)
	for i, v := range fields {
		if v == "namespace" {
			v = "ns"
		}
		headings[i] = fmt.Sprintf(pad, v)
	}
	return "  - " + strings.Join(headings, " : %v\n    ") + " : %v\n"
}

func (y *yamlWriter) Head(path string, scanned, created time.Time, version [3]int, ids [][2]string, fields [][]string, hh string) {
	y.hh = hh
	y.hstrs = make([]string, len(fields))
	y.vals = make([][]interface{}, len(fields))
	for i, f := range fields {
		y.hstrs[i] = header(f)
		y.vals[i] = make([]interface{}, len(f))
	}
	fmt.Fprintf(y.w,
		"---\nsiegfried   : %d.%d.%d\nscandate    : %v\nsignature   : %s\ncreated     : %v\nidentifiers : \n",
		version[0], version[1], version[2],
		scanned.Format(time.RFC3339),
		y.replacer.Replace(path),
		created.Format(time.RFC3339))
	for _, id := range ids {
		fmt.Fprintf(y.w, "  - name    : '%v'\n    details : '%v'\n", id[0], id[1])
	}
}

func (y *yamlWriter) File(name string, sz int64, mod string, checksum []byte, err error, ids []core.Identification) {
	var (
		errStr   string
		h        string
		thisName string
		idx      int = -1
	)
	if err != nil {
		errStr = "'" + y.replacer.Replace(err.Error()) + "'"
	}
	if checksum != nil {
		h = fmt.Sprintf("%-8s : %s\n", y.hh, hex.EncodeToString(checksum))
	}
	fmt.Fprintf(y.w, "---\nfilename : '%s'\nfilesize : %d\nmodified : %s\nerrors   : %s\n%smatches  :\n", y.replacer.Replace(name), sz, mod, errStr, h)
	for _, id := range ids {
		values := id.Values()
		if values[0] != thisName {
			idx++
			thisName = values[0]
		}
		for i, v := range values {
			if v == "" {
				y.vals[idx][i] = ""
				continue
			}
			y.vals[idx][i] = "'" + y.replacer.Replace(v) + "'"
		}
		fmt.Fprintf(y.w, y.hstrs[idx], y.vals[idx]...)
	}
}

func (y *yamlWriter) Tail() { y.w.Flush() }

type jsonWriter struct {
	subs     bool
	replacer *strings.Replacer
	w        *bufio.Writer
	hh       string
	hstrs    []func([]string) string
}

func JSON(w io.Writer) Writer {
	return &jsonWriter{
		replacer: strings.NewReplacer(
			`\`, `\\`,
			`"`, `\"`,
			"\u0000", `\u0000`,
			"\u0001", `\u0001`,
			"\u0002", `\u0002`,
			"\u0003", `\u0003`,
			"\u0004", `\u0004`,
			"\u0005", `\u0005`,
			"\u0006", `\u0006`,
			"\u0007", `\u0007`,
			"\u0008", `\u0008`,
			"\u0009", `\u0009`,
			"\u000A", `\u000A`,
			"\u000B", `\u000B`,
			"\u000C", `\u000C`,
			"\u000D", `\u000D`,
			"\u000E", `\u000E`,
			"\u000F", `\u000F`,
			"\u0010", `\u0010`,
			"\u0011", `\u0011`,
			"\u0012", `\u0012`,
			"\u0013", `\u0013`,
			"\u0014", `\u0014`,
			"\u0015", `\u0015`,
			"\u0016", `\u0016`,
			"\u0017", `\u0017`,
			"\u0018", `\u0018`,
			"\u0019", `\u0019`,
		),
		w: bufio.NewWriter(w),
	}
}

func jsonizer(fields []string) func([]string) string {
	for i, v := range fields {
		if v == "namespace" {
			fields[i] = "\"ns\":\""
			continue
		}
		fields[i] = "\"" + v + "\":\""
	}
	vals := make([]string, len(fields))
	return func(values []string) string {
		for i, v := range values {
			vals[i] = fields[i] + v
		}
		return "{" + strings.Join(vals, "\",") + "\"}"
	}
}

func (j *jsonWriter) Head(path string, scanned, created time.Time, version [3]int, ids [][2]string, fields [][]string, hh string) {
	j.hh = hh
	j.hstrs = make([]func([]string) string, len(fields))
	for i, f := range fields {
		j.hstrs[i] = jsonizer(f)
	}
	fmt.Fprintf(j.w,
		"{\"siegfried\":\"%d.%d.%d\",\"scandate\":\"%v\",\"signature\":\"%s\",\"created\":\"%v\",\"identifiers\":[",
		version[0], version[1], version[2],
		scanned.Format(time.RFC3339),
		path,
		created.Format(time.RFC3339))
	for i, id := range ids {
		if i > 0 {
			j.w.WriteString(",")
		}
		fmt.Fprintf(j.w, "{\"name\":\"%s\",\"details\":\"%s\"}", id[0], id[1])
	}
	j.w.WriteString("],\"files\":[")
}

func (j *jsonWriter) File(name string, sz int64, mod string, checksum []byte, err error, ids []core.Identification) {
	if j.subs {
		j.w.WriteString(",")
	}
	var (
		errStr   string
		h        string
		thisName string
		idx      int = -1
	)
	if err != nil {
		errStr = err.Error()
	}
	if checksum != nil {
		h = fmt.Sprintf("\"%s\":\"%s\",", j.hh, hex.EncodeToString(checksum))
	}
	fmt.Fprintf(j.w, "{\"filename\":\"%s\",\"filesize\": %d,\"modified\":\"%s\",\"errors\": \"%s\",%s\"matches\": [", j.replacer.Replace(name), sz, mod, errStr, h)
	for i, id := range ids {
		if i > 0 {
			j.w.WriteString(",")
		}
		values := id.Values()
		if values[0] != thisName {
			idx++
			thisName = values[0]
		}
		j.w.WriteString(j.hstrs[idx](values))
	}
	j.w.WriteString("]}")
	j.subs = true
}

func (j *jsonWriter) Tail() {
	j.w.WriteString("]}\n")
	j.w.Flush()
}

type droidWriter struct {
	id      int
	parents map[string]parent
	rec     []string
	w       *csv.Writer
}

type parent struct {
	id      int
	uri     string
	archive string
}

func Droid(w io.Writer) Writer {
	return &droidWriter{
		parents: make(map[string]parent),
		rec:     make([]string, 18),
		w:       csv.NewWriter(w),
	}
}

// "identifier", "id", "format name", "format version", "mimetype", "basis", "warning"
func (d *droidWriter) Head(path string, scanned, created time.Time, version [3]int, ids [][2]string, fields [][]string, hh string) {
	if hh == "" {
		hh = "no"
	}
	d.w.Write([]string{
		"ID", "PARENT_ID", "URI", "FILE_PATH", "NAME",
		"METHOD", "STATUS", "SIZE", "TYPE", "EXT",
		"LAST_MODIFIED", "EXTENSION_MISMATCH", strings.ToUpper(hh) + "_HASH", "FORMAT_COUNT",
		"PUID", "MIME_TYPE", "FORMAT_NAME", "FORMAT_VERSION"})
}

func (d *droidWriter) File(p string, sz int64, mod string, checksum []byte, err error, ids []core.Identification) {
	d.id++
	d.rec[0], d.rec[6], d.rec[10] = strconv.Itoa(d.id), "Done", mod
	if err != nil {
		d.rec[6] = err.Error()
	}
	d.rec[1], d.rec[2], d.rec[3], d.rec[4], d.rec[9] = d.processPath(p)
	// if folder (has sz -1) or error
	if sz < 0 || ids == nil {
		d.rec[5], d.rec[7], d.rec[12], d.rec[13], d.rec[14], d.rec[15], d.rec[16], d.rec[17] = "", "", "", "", "", "", "", ""
		if sz < 0 {
			d.rec[8], d.rec[9], d.rec[11] = "Folder", "", "false"
			d.parents[d.rec[3]] = parent{d.id, d.rec[2], ""}
		} else {
			d.rec[8], d.rec[11] = "", ""
		}
		d.rec[3] = clearArchivePath(d.rec[2], d.rec[3])
		d.w.Write(d.rec)
		return
	}
	// size
	d.rec[7] = strconv.FormatInt(sz, 10)
	if checksum == nil {
		d.rec[12] = ""
	} else {
		d.rec[12] = hex.EncodeToString(checksum)
	}
	// leave early for unknowns
	if len(ids) < 1 || !ids[0].Known() {
		d.rec[5], d.rec[8], d.rec[11], d.rec[13] = "", "File", "FALSE", "0"
		d.rec[14], d.rec[15], d.rec[16], d.rec[17] = "", "", "", ""
		d.rec[3] = clearArchivePath(d.rec[2], d.rec[3])
		d.w.Write(d.rec)
		return
	}
	d.rec[13] = strconv.Itoa(len(ids))
	for _, id := range ids {
		if id.Archive() > config.None {
			d.rec[8] = "Container"
			d.parents[d.rec[3]] = parent{d.id, d.rec[2], id.Archive().String()}
		} else {
			d.rec[8] = "File"
		}
		fields := id.Values()
		d.rec[5], d.rec[11] = getMethod(fields[5]), mismatch(fields[6])
		d.rec[14], d.rec[15], d.rec[16], d.rec[17] = fields[1], fields[4], fields[2], fields[3]
		d.rec[3] = clearArchivePath(d.rec[2], d.rec[3])
		d.w.Write(d.rec)
	}
}

func (d *droidWriter) Tail() { d.w.Flush() }

func (d *droidWriter) processPath(p string) (parent, uri, path, name, ext string) {
	path, _ = filepath.Abs(p)
	path = strings.TrimSuffix(path, string(filepath.Separator))
	name = filepath.Base(path)
	dir := filepath.Dir(path)
	par, ok := d.parents[dir]
	if ok {
		parent = strconv.Itoa(par.id)
		uri = toUri(par.uri, par.archive, escape(name))
	} else {
		puri := "file:/" + escape(filepath.ToSlash(dir))
		uri = toUri(puri, "", escape(name))
	}
	ext = strings.TrimPrefix(filepath.Ext(p), ".")
	return
}

func toUri(parenturi, parentarc, base string) string {
	if len(parentarc) > 0 {
		parenturi = parentarc + ":" + parenturi + "!"
	}
	return parenturi + "/" + base
}

// uri escaping adapted from https://golang.org/src/net/url/url.go
func shouldEscape(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '_', '.', '~', '/', ':':
		return false
	}
	return true
}

func escape(s string) string {
	var hexCount int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}
	if hexCount == 0 {
		return s
	}
	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		if c := s[i]; shouldEscape(c) {
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		} else {
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func clearArchivePath(uri, path string) string {
	if strings.HasPrefix(uri, config.Zip.String()) ||
		strings.HasPrefix(uri, config.Tar.String()) ||
		strings.HasPrefix(uri, config.Gzip.String()) {
		path = ""
	}
	return path
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
