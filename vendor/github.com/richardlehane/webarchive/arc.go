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

package webarchive

import (
	"bytes"
	"io"
	"strconv"
	"time"
)

// ARCTime is a time format string for the ARC time format
const ARCTime = "20060102150405"

// ARCRecord represents the common fields shared by ARC version 1
// and ARC version 2 URL record blocks.
// ARC version 2 URL record blocks have additional fields not exposed
// here. These fields are available in the Fields() map.
// To access the IP() method of an ARCRecord, do an interface
// assertion on a Record.
//
// Example:
//  record, _ := reader.Next()
//  arcrecord, ok := record.(ARCRecord)
//  if ok {fmt.Println(arcrecord.IP())}
type ARCRecord interface {
	IP() string
	Record
}

// ARC structs represent the Version blocks at the start of ARC files.
// Provides information about the ARC file as a whole such as version,
// file path of the archive file, and date creation of the archive file.
type ARC struct {
	FileDesc   string    // Original pathname of the archive file
	Address    string    // IP address of machine that created the archive file
	FileDate   time.Time // Date the archive file was created
	Version    int       // ARC version (1 or 2) - this will affect the fields available in the Fields() map
	OriginCode string    // Name of gathering organization
}

// ARCReader is the ARC implementation of a webarchive Reader
type ARCReader struct {
	*ARC
	*reader
	arcHeader
}

type arcHeader interface {
	IP() string
	Header
	size() int64
	setfields([]byte)
}

// Version 1 URL record
type url1 struct {
	url    string
	ip     string    // dotted-quad (eg 192.216.46.98 or 0.0.0.0)
	date   time.Time //  YYYYMMDDhhmmss (Greenwich Mean Time)
	mime   string    // "no-type"|MIME type of data (e.g., "text/html")
	sz     int64
	fields []byte
}

func (u *url1) URL() string     { return u.url }
func (u *url1) Date() time.Time { return u.date }
func (u *url1) Fields() map[string][]string {
	var fields map[string][]string
	if len(u.fields) > 0 {
		fields = getAllValues(u.fields)
	} else {
		fields = make(map[string][]string)
	}
	fields["URL"] = []string{u.url}
	fields["IP"] = []string{u.ip}
	fields["Date"] = []string{u.date.Format(ARCTime)}
	fields["MIME"] = []string{u.mime}
	fields["Size"] = []string{strconv.FormatInt(u.sz, 10)}
	return fields
}

func (u *url1) IP() string   { return u.ip }
func (u *url1) MIME() string { return u.mime }

func (u *url1) transferEncodings() []string {
	if len(u.fields) == 0 {
		return nil
	}
	vals := getSelectValues(u.fields, "Transfer-Encoding")
	if vals[0] == "" {
		return nil
	}
	return splitAndReverse(vals[0])
}
func (u *url1) encodings() []string {
	if len(u.fields) == 0 {
		return nil
	}
	vals := getSelectValues(u.fields, "Transfer-Encoding", "Content-Encoding")
	if vals[0] == "" {
		if vals[1] == "" {
			return nil
		}
		return splitAndReverse(vals[1])
	}
	if vals[1] == "" {
		return splitAndReverse(vals[0])
	}
	return append(splitAndReverse(vals[0]), splitAndReverse(vals[1])...)
}

func (u *url1) size() int64        { return u.sz }
func (u *url1) setfields(f []byte) { u.fields = f }

// Version 2 URL record
type url2 struct {
	*url1
	statusCode int
	checksum   string
	location   string
	offset     int64
	filename   string
}

func (u *url2) Fields() map[string][]string {
	fields := u.url1.Fields()
	fields["StatusCode"] = []string{strconv.Itoa(u.statusCode)}
	fields["Checksum"] = []string{u.checksum}
	fields["Location"] = []string{u.location}
	fields["Offset"] = []string{strconv.FormatInt(u.offset, 10)}
	fields["Filename"] = []string{u.location}
	return fields
}

// NewARCReader creates a new ARC reader from the supplied io.Reader.
// Use instead of NewReader if you are only working with ARC files.
func NewARCReader(r io.Reader) (*ARCReader, error) {
	rdr, err := newReader(r)
	if err != nil {
		return nil, err
	}
	return newARCReader(rdr)
}

func newARCReader(r *reader) (*ARCReader, error) {
	arc := &ARCReader{reader: r}
	var err error
	arc.ARC, err = arc.readVersionBlock()
	return arc, err
}

// Reset allows re-use of an ARC reader
func (a *ARCReader) Reset(r io.Reader) error {
	a.reader.reset(r)
	return a.reset()
}

func (a *ARCReader) reset() error {
	var err error
	a.ARC, err = a.readVersionBlock()
	return err
}

// Next iterates to the next Record. Returns io.EOF at the end of file.
func (a *ARCReader) Next() (Record, error) {
	buf, err := a.next()
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(bytes.TrimSpace(buf), []byte(" "))
	if a.Version == 1 {
		a.arcHeader, err = makeUrl1(parts)
	} else {
		a.arcHeader, err = makeUrl2(parts)
	}
	if err != nil {
		return nil, err
	}
	a.thisIdx, a.sz = 0, a.size()
	return a, err
}

// NextPayload iterates to the next payload record.
// As ARC files do not differentiate between different types of records,
// the effect of NextPayload for an ARC reader is just to strip HTTP
// headers. These stripped headers are then made available in the Fields() map.
func (a *ARCReader) NextPayload() (Record, error) {
	r, err := a.Next()
	if err != nil {
		return r, err
	}
	if v, err := a.peek(5); err == nil && string(v) == "HTTP/" {
		f, err := a.storeLines(0, true)
		if err != nil {
			return r, err
		}
		a.setfields(f)
	}
	return r, err
}

func (r *ARCReader) readVersionBlock() (*ARC, error) {
	buf, _ := r.readLine()
	if len(buf) == 0 {
		return nil, ErrVersionBlock
	}
	line1 := bytes.Split(buf, []byte(" "))
	if len(line1) < 3 {
		return nil, ErrVersionBlock
	}
	t, err := time.Parse(ARCTime, string(line1[2]))
	if err != nil {
		return nil, ErrVersionBlock
	}
	buf, _ = r.readLine()
	line2 := bytes.Split(buf, []byte(" "))
	if len(line2) < 3 {
		return nil, ErrVersionBlock
	}
	version, err := strconv.Atoi(string(line2[0]))
	if err != nil {
		return nil, ErrVersionBlock
	}
	l, err := strconv.Atoi(string(bytes.TrimSpace(line1[len(line1)-1])))
	if err != nil {
		return nil, ErrVersionBlock
	}
	// now scan ahead to first doc
	l -= len(buf)
	if r.slicer {
		r.idx += int64(l)
	} else {
		discard(r.buf, l)
	}
	return &ARC{
		FileDesc:   string(line1[0]),
		Address:    string(line1[1]),
		FileDate:   t,
		Version:    version,
		OriginCode: string(bytes.TrimSpace(line2[len(line2)-1])),
	}, nil
}

func makeUrl1(p [][]byte) (*url1, error) {
	if len(p) < 5 {
		return nil, ErrARCHeader
	}
	date, err := time.Parse(ARCTime, string(p[2]))
	if err != nil {
		return nil, ErrARCHeader
	}
	l, err := strconv.ParseInt(string(p[len(p)-1]), 10, 64)
	if err != nil {
		return nil, ErrARCHeader
	}
	return &url1{
		url:  string(p[0]),
		ip:   string(p[1]),
		date: date,
		mime: string(p[3]),
		sz:   l,
	}, nil
}

func makeUrl2(p [][]byte) (*url2, error) {
	if len(p) != 10 {
		return nil, ErrARCHeader
	}
	u1, err := makeUrl1(p)
	if err != nil {
		return nil, ErrARCHeader
	}
	status, err := strconv.Atoi(string(p[4]))
	if err != nil {
		return nil, ErrARCHeader
	}
	offset, err := strconv.ParseInt(string(p[7]), 10, 64)
	if err != nil {
		return nil, ErrARCHeader
	}
	return &url2{
		url1:       u1,
		statusCode: status,
		checksum:   string(p[5]),
		location:   string(p[6]),
		offset:     offset,
		filename:   string(p[8]),
	}, nil
}
