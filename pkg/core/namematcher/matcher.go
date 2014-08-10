package namematcher

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"sort"
)

type Matcher interface {
	Identify(string) []int
	String() string
	Save(io.Writer) (int, error)
}

func Load(r io.Reader) (Matcher, error) {
	e := NewExtensionMatcher()
	dec := gob.NewDecoder(r)
	err := dec.Decode(&e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e ExtensionMatcher) Save(w io.Writer) (int, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(e)
	if err != nil {
		return 0, err
	}
	sz := buf.Len()
	_, err = buf.WriteTo(w)
	if err != nil {
		return 0, err
	}
	return sz, nil
}

func (e ExtensionMatcher) String() string {
	var str string
	var keys []string
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, v := range keys {
		str += fmt.Sprintf("%v: %v\n", v, e[v])
	}
	return str
}
