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
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type sfYAML struct {
	buf    *bufio.Reader
	closer io.ReadCloser
}

const (
	record = iota
	keyval
	item // e.g. -
)

type token struct {
	typ int
	key string
	val string
}

func advance(buf *bufio.Reader) (token, error) {
	byts, err := buf.ReadBytes('\n')
	if err != nil {
		return token{}, err
	}
	if bytes.Equal(byts, []byte("---\n")) {
		return token{typ: record}, nil
	}
	var tok token
	if bytes.HasPrefix(byts, []byte("  - ")) {
		tok.typ = item
		byts = byts[4:]
	} else {
		tok.typ = keyval
	}
	split := bytes.SplitN(byts, []byte(":"), 1)
	tok.key = string(bytes.TrimSpace(split[0]))
	if len(split) == 2 {
		tok.val = string(bytes.TrimSuffix(bytes.TrimPrefix(bytes.TrimSpace(split[1]), []byte("'")), []byte("'")))
	}
	return tok, nil
}

func newYAML(rc io.ReadCloser, path string) (Reader, error) {
	buf := bufio.NewReader(rc)
	sfy := &sfYAML{
		buf:    buf,
		closer: rc,
	}
	tok, err := advance(buf)
	if err != nil || tok.typ != record {
		return nil, fmt.Errorf("invalid YAML; got %v", err)
	}
	return sfy, nil
}

func (sfy *sfYAML) Head() Head {
	return Head{}
}

func (sfy *sfYAML) Next() (File, error) {
	return File{}, io.EOF
}

func (sfy *sfYAML) Close() error {
	return sfy.closer.Close()
}
