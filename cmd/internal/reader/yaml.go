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
	"strings"
)

type sfYAML struct {
	replacer *strings.Replacer
	buf      *bufio.Reader
	closer   io.ReadCloser
	head     Head
	peek     record
	err      error
}

const (
	divide = iota
	keyval
	item // e.g. -
)

type token struct {
	typ int
	key string
	val string
}

func advance(buf *bufio.Reader, repl *strings.Replacer) (token, error) {
	byts, err := buf.ReadBytes('\n')
	if err != nil {
		return token{}, err
	}
	if bytes.Equal(byts, []byte("---\n")) {
		return token{typ: divide}, nil
	}
	var tok token
	if bytes.HasPrefix(byts, []byte("  - ")) {
		tok.typ = item
		byts = byts[4:]
	} else {
		tok.typ = keyval
	}
	split := bytes.SplitN(byts, []byte(":"), 2)
	tok.key = string(bytes.TrimSpace(split[0]))
	if len(split) == 2 {
		tok.val = repl.Replace(string(bytes.TrimSuffix(bytes.TrimPrefix(bytes.TrimSpace(split[1]), []byte("'")), []byte("'"))))
	}
	return tok, nil
}

func consumeList(buf *bufio.Reader, repl *strings.Replacer, tok token) ([]string, []string, error) {
	fields, values := []string{tok.key}, []string{tok.val}
	var err error
	for tok, err = advance(buf, repl); err == nil && tok.typ != divide; tok, err = advance(buf, repl) {
		fields, values = append(fields, tok.key), append(values, tok.val)
	}
	return fields, values, err
}

func consumeRecord(buf *bufio.Reader, repl *strings.Replacer) (record, error) {
	var (
		rec record
		tok token
		err error
	)
	m := make(map[string]string)
	for tok, err = advance(buf, repl); err == nil && tok.typ == keyval; tok, err = advance(buf, repl) {
		m[tok.key] = tok.val
	}
	if err != nil || tok.typ != item {
		if err == nil {
			return rec, fmt.Errorf("unexpected token got %s", tok.typ)
		}
		return rec, err
	}
	ks, vs, err := consumeList(buf, repl, tok)
	if err != nil && err != io.EOF {
		return rec, err
	}
	return record{m, ks, vs}, nil
}

func newYAML(rc io.ReadCloser, path string) (Reader, error) {
	sfy := &sfYAML{
		replacer: strings.NewReplacer("''", "'"),
		buf:      bufio.NewReader(rc),
		closer:   rc,
	}
	return sfy.makeHead(path)
}

func (sfy *sfYAML) Head() Head {
	return sfy.head
}

func (sfy *sfYAML) makeHead(path string) (*sfYAML, error) {
	tok, err := advance(sfy.buf, sfy.replacer)
	if err != nil || tok.typ != divide {
		return nil, fmt.Errorf("invalid YAML; got %v", err)
	}
	rec, err := consumeRecord(sfy.buf, sfy.replacer)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML; got %v", err)
	}
	rec.attributes["results"] = path
	sfy.head, err = getHead(rec)
	sfy.peek, sfy.err = consumeRecord(sfy.buf, sfy.replacer)
	sfy.head.HashHeader = getHash(sfy.peek.attributes)
	sfy.head.Fields = getFields(sfy.peek.listFields, sfy.peek.listValues)
	return sfy, err
}

func (sfy *sfYAML) Next() (File, error) {
	r, e := sfy.peek, sfy.err
	if e != nil {
		return File{}, e
	}
	sfy.peek, sfy.err = consumeRecord(sfy.buf, sfy.replacer)
	return getFile(r)
}

func (sfy *sfYAML) Close() error {
	return sfy.closer.Close()
}
