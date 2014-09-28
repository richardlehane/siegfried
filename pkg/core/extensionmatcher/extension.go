package extensionmatcher

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type Matcher map[string][]Result

func New() Matcher {
	return make(Matcher)
}

type Result int

func (r Result) Index() int {
	return int(r)
}

func (r Result) Basis() string {
	return "extension match"
}

func (e Matcher) Add(ext string, fmt int) {
	_, ok := e[ext]
	if ok {
		e[ext] = append(e[ext], Result(fmt))
		return
	}
	e[ext] = []Result{Result(fmt)}
}

func (e Matcher) Identify(name string, na *siegreader.Buffer) chan core.Result {
	res := make(chan core.Result, 10)
	go func() {
		ext := filepath.Ext(name)
		if len(ext) > 0 {
			fmts, ok := e[strings.TrimPrefix(ext, ".")]
			if ok {
				for _, v := range fmts {
					res <- v
				}
			}
		}
		close(res)
	}()
	return res
}

func Load(r io.Reader) (core.Matcher, error) {
	e := New()
	dec := gob.NewDecoder(r)
	err := dec.Decode(&e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e Matcher) Save(w io.Writer) (int, error) {
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

func (e Matcher) String() string {
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

func (e Matcher) Priority() bool {
	return false
}
