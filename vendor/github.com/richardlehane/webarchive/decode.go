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
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http/httputil"
)

func isgzip(buf []byte) bool {
	if buf[0] != 0x1f || buf[1] != 0x8b || buf[2] != 8 {
		return false
	}
	return true
}

const zlibDeflate = 8

func iszlib(buf []byte) bool {
	h := uint(buf[0])<<8 | uint(buf[1])
	if (buf[0]&0x0f != zlibDeflate) || (h%31 != 0) {
		return false
	}
	return true
}

func ischunk(buf []byte) bool {
	for i, c := range buf {
		switch {
		case '0' >= c && c <= '9':
			continue
		case 'a' <= c && c <= 'f':
			continue
		case 'A' <= c && c <= 'F':
			continue
		case c == '\r':
			if i > 0 && i < len(buf)-1 && buf[i+1] == '\n' {
				return true
			}
			return false
		default:
			return false
		}
	}
	return false
}

type payloadDecoder struct {
	Record
	rdr io.Reader
}

func (pd *payloadDecoder) Read(b []byte) (int, error) {
	return pd.rdr.Read(b)
}

func (pd *payloadDecoder) IsSlicer() bool {
	return false
}

func newDecoder(rec Record, encodings []string) Record {
	if len(encodings) == 0 {
		return rec
	}
	pd := &payloadDecoder{Record: rec, rdr: rec}
	for i, v := range encodings {
		switch v {
		case "chunked":
			if i == 0 {
				if peek, err := rec.peek(10); err != nil || !ischunk(peek) {
					return rec
				}
			}
			pd.rdr = httputil.NewChunkedReader(pd.rdr)
		case "deflate":
			if i == 0 {
				if peek, err := rec.peek(2); err != nil || !iszlib(peek) {
					return rec
				}
			}
			rdr, err := zlib.NewReader(pd.rdr)
			if err == nil {
				pd.rdr = rdr
			}
		case "gzip":
			if i == 0 {
				if peek, err := rec.peek(3); err != nil || !isgzip(peek) {
					return rec
				}
			}
			rdr, err := gzip.NewReader(pd.rdr)
			if err == nil {
				pd.rdr = rdr
			}
		}
	}
	return pd
}

// DecodePayload decodes any encodings (transfer or content) declared in a record's HTTP header.
// Decodes chunked, deflate and gzip encodings.
func DecodePayload(r Record) Record {
	return newDecoder(r, r.encodings())
}

// DecodePayloadT decodes any transfer encodings declared in a record's HTTP header.
// Decodes chunked, deflate and gzip encodings.
func DecodePayloadT(r Record) Record {
	return newDecoder(r, r.transferEncodings())
}
