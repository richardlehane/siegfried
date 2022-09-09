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
	replacer    *strings.Replacer
	dblReplacer *strings.Replacer
	buf         *bufio.Reader
	head        Head
	peek        record
	err         error
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

func advance(buf *bufio.Reader, repl, dbl *strings.Replacer) (token, error) {
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
		val := bytes.TrimSpace(split[1])
		if len(val) > 0 {
			if val[0] == '"' {
				tok.val = dbl.Replace(string(bytes.TrimSuffix(bytes.TrimPrefix(val, []byte("\"")), []byte("\""))))
			} else {
				tok.val = repl.Replace(string(bytes.TrimSuffix(bytes.TrimPrefix(val, []byte("'")), []byte("'"))))
			}
		}
	}
	return tok, nil
}

func consumeList(buf *bufio.Reader, repl, dbl *strings.Replacer, tok token) ([]string, []string, error) {
	fields, values := []string{tok.key}, []string{tok.val}
	var err error
	for tok, err = advance(buf, repl, dbl); err == nil && tok.typ != divide; tok, err = advance(buf, repl, dbl) {
		fields, values = append(fields, tok.key), append(values, tok.val)
	}
	return fields, values, err
}

func consumeRecord(buf *bufio.Reader, repl, dbl *strings.Replacer) (record, error) {
	var (
		rec record
		tok token
		err error
	)
	m := make(map[string]string)
	for tok, err = advance(buf, repl, dbl); err == nil && tok.typ == keyval; tok, err = advance(buf, repl, dbl) {
		m[tok.key] = tok.val
	}
	if err != nil || tok.typ != item {
		if err == nil {
			return rec, fmt.Errorf("unexpected token got %d", tok.typ)
		}
		return rec, err
	}
	ks, vs, err := consumeList(buf, repl, dbl, tok)
	if err != nil && err != io.EOF {
		return rec, err
	}
	return record{m, ks, vs}, nil
}

func newYAML(r io.Reader, path string) (Reader, error) {
	sfy := &sfYAML{
		replacer: strings.NewReplacer("''", "'"),
		dblReplacer: strings.NewReplacer(
			"\\0", "\x00",
			"\\a", "\x07",
			"\\b", "\x08",
			"\\n", "\x0A",
			"\\v", "\x0B",
			"\\f", "\x0C",
			"\\r", "\x0D",
			"\\e", "\x1B",
			"\\\"", "\x22",
			"\\/", "\x2F",
			"\\\\", "\x5c",
		),
		buf: bufio.NewReader(r),
	}
	return sfy.makeHead(path)
}

func (sfy *sfYAML) Head() Head {
	return sfy.head
}

func (sfy *sfYAML) makeHead(path string) (*sfYAML, error) {
	tok, err := advance(sfy.buf, sfy.replacer, sfy.dblReplacer)
	if err != nil || tok.typ != divide {
		return nil, fmt.Errorf("invalid YAML; got %v", err)
	}
	rec, err := consumeRecord(sfy.buf, sfy.replacer, sfy.dblReplacer)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML; got %v", err)
	}
	rec.attributes["results"] = path
	sfy.head, err = getHead(rec)
	sfy.peek, sfy.err = consumeRecord(sfy.buf, sfy.replacer, sfy.dblReplacer)
	sfy.head.HashHeader = getHash(sfy.peek.attributes)
	sfy.head.Fields = getFields(sfy.peek.listFields, sfy.peek.listValues)
	return sfy, err
}

func (sfy *sfYAML) Next() (File, error) {
	r, e := sfy.peek, sfy.err
	if e != nil {
		return File{}, e
	}
	sfy.peek, sfy.err = consumeRecord(sfy.buf, sfy.replacer, sfy.dblReplacer)
	return getFile(r)
}
