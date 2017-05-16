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
	"encoding/json"
	"io"
	"strconv"
)

type sfJSON struct {
	dec  *json.Decoder
	head Head
	peek record
	err  error
}

func next(dec *json.Decoder) ([]string, []string, error) {
	var (
		tok json.Token
		err error
		i   int
	)
	keys, vals := make([]string, 0, 10), make([]string, 0, 10)
	for tok, err = dec.Token(); err == nil; tok, err = dec.Token() {
		switch tok := tok.(type) {
		case string:
			if i%2 == 0 {
				keys = append(keys, tok)
			} else {
				vals = append(vals, tok)
			}
			i++
		case float64:
			i++
			vals = append(vals, strconv.FormatFloat(tok, 'f', 0, 32))
		case json.Delim:
			if tok.String() == "[" || tok.String() == "]" {
				return keys, vals, nil
			}
		}
	}
	return nil, nil, err
}

func jsonRecord(dec *json.Decoder) (record, error) {
	keys, vals, err := next(dec)
	if err != nil {
		return record{}, err
	}
	m := make(map[string]string)
	for i, v := range vals {
		m[keys[i]] = v
	}
	keys, vals, err = next(dec)
	if err != nil {
		return record{}, err
	}
	return record{m, keys, vals}, nil
}

func newJSON(r io.Reader, path string) (Reader, error) {
	sfj := &sfJSON{dec: json.NewDecoder(r)}
	rec, err := jsonRecord(sfj.dec)
	if err != nil {
		return nil, err
	}
	rec.attributes["results"] = path
	sfj.head, err = getHead(rec)
	if err != nil {
		return nil, err
	}
	next(sfj.dec) // throw away "files": [
	sfj.peek, sfj.err = jsonRecord(sfj.dec)
	sfj.head.HashHeader = getHash(sfj.peek.attributes)
	sfj.head.Fields = getFields(sfj.peek.listFields, sfj.peek.listValues)
	return sfj, nil
}

func (sfj *sfJSON) Head() Head {
	return sfj.head
}

func (sfj *sfJSON) Next() (File, error) {
	r, e := sfj.peek, sfj.err
	if e != nil {
		return File{}, e
	}
	sfj.peek, sfj.err = jsonRecord(sfj.dec)
	return getFile(r)
}
