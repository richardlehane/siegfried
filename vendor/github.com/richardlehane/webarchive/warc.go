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
	"io"
	"strconv"
	"time"
)

// WARCRecord allows access to specific WARC record fields. Other WARC
// fields not included here are accessible via the Fields() method.
// To access the ID() and Type() methods of a WARCRecord, do an interface
// assertion on a Record.
//
// Example:
//  record, _ := reader.Next()
//  warcrecord, ok := record.(WARCRecord)
//  if ok {fmt.Println(warcrecord.ID())}
type WARCRecord interface {
	ID() string
	Type() string
	Record
}

type warcHeader struct {
	url     string    // WARC-Target-URI
	id      string    // WARC-Record-ID
	date    time.Time // WARC-Date
	typ     string    // WARC-Type
	segment int       // WARC-Segment-Number
	mime    string    // WARC-Identified-Payload-Type or HTTP Content-Type header
	fields  []byte
}

// URL returns the URL of the current Record.
func (h *warcHeader) URL() string { return h.url }

// Date returns the archive date of the current Record.
func (h *warcHeader) Date() time.Time { return h.date }

func (h *warcHeader) MIME() string {
	if h.mime != "" {
		return h.mime
	}
	ctypes := getSingleValues(h.fields, "Content-Type")
	switch len(ctypes) {
	case 0:
		return ""
	case 1:
		return ctypes[0]
	default:
		return ctypes[1]
	}
}

func (h *warcHeader) transferEncodings() []string {
	vals := getSelectValues(h.fields, "Transfer-Encoding")
	if vals[0] == "" {
		return nil
	}
	return splitAndReverse(vals[0])
}
func (h *warcHeader) encodings() []string {
	vals := getSelectValues(h.fields, "Transfer-Encoding", "Content-Encoding")
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

// Fields returns a map of all WARC fields for the current Record.
// If NextPayload was used, this map will also contain any stripped HTTP headers.
func (h *warcHeader) Fields() map[string][]string { return getAllValues(h.fields) }

// ID returns the WARC Record ID.
func (h *warcHeader) ID() string { return h.id }

// Type returns the WARC Type
func (h *warcHeader) Type() string { return h.typ }

// WARCReader is the WARC implementation of a webarchive Reader
type WARCReader struct {
	*warcHeader
	*reader
	continuations
}

// NewWARCReader creates a new WARC reader from the supplied io.Reader.
// Use instead of NewReader if you are only working with ARC files.
func NewWARCReader(r io.Reader) (*WARCReader, error) {
	rdr, err := newReader(r)
	if err != nil {
		return nil, err
	}
	return newWARCReader(rdr)
}

func newWARCReader(r *reader) (*WARCReader, error) {
	w := &WARCReader{&warcHeader{}, r, nil}
	return w, w.reset()
}

// Reset allows re-use of a ARC reader
func (w *WARCReader) Reset(r io.Reader) error {
	w.reader.reset(r)
	return w.reset()
}

func (w *WARCReader) reset() error {
	if v, err := w.peek(4); err != nil || string(v) != "WARC" {
		return ErrWARCHeader
	}
	return nil
}

// Next iterates to the next Record. Returns io.EOF at the end of file.
func (w *WARCReader) Next() (Record, error) {
	// discard the returned slice as the first line in a WARC record is just the WARC header
	_, err := w.next()
	if err != nil {
		return nil, err
	}
	w.fields, err = w.storeLines(0, false)
	if err != nil {
		return nil, ErrWARCRecord
	}
	vals := getSelectValues(w.fields, "WARC-Type", "WARC-Target-URI", "WARC-Date", "Content-Length", "WARC-Record-ID", "WARC-Segment-Number", "WARC-Identified-Payload-Type")
	w.typ, w.url, w.id, w.mime = vals[0], vals[1], vals[4], vals[6]
	w.date, err = time.Parse(time.RFC3339, vals[2])
	if err != nil {
		return nil, err
	}
	w.sz, err = strconv.ParseInt(vals[3], 10, 64)
	if err != nil {
		return nil, err
	}
	w.thisIdx = 0
	if vals[5] != "" {
		w.segment, err = strconv.Atoi(vals[5])
		if err != nil {
			return nil, err
		}
	} else {
		w.segment = 0
	}
	return w, nil
}

// NextPayload iterates to the next payload record.
// It skips non-resource, conversion or response records and merges continuations into single records.
// It also strips HTTP headers from response records. After stripping, those HTTP headers are available alongside
// the WARC headers in the record.Fields() map.
func (w *WARCReader) NextPayload() (Record, error) {
	for {
		r, err := w.Next()
		if err != nil {
			return r, err
		}
		if w.segment > 0 {
			if w.continuations == nil {
				w.continuations = make(continuations)
			}
			if c, ok := w.continuations.put(w); ok {
				return c, nil
			}
			continue
		}
		switch w.typ {
		default:
			continue
		case "resource", "conversion":
			return r, err
		case "response":
			if v, err := w.peek(5); err == nil && string(v) == "HTTP/" {
				l := len(w.fields)
				w.fields, err = w.storeLines(l, true)
			}
			return r, err
		}
	}
}
