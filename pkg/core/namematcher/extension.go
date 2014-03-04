package namematcher

import (
	"path/filepath"

	"github.com/richardlehane/siegfried/pkg/core"
)

type ExtensionMatcher struct {
	m    map[string][]int
	name string
}

func NewExtensionMatcher() *ExtensionMatcher {
	return &ExtensionMatcher{m: make(map[string][]int)}
}

func (e *ExtensionMatcher) Add(ext string, fmt int) {
	_, ok := e.m[ext]
	if ok {
		e.m[ext] = append(e.m[ext], fmt)
		return
	}
	e.m[ext] = []int{fmt}
}

func (e *ExtensionMatcher) SetName(name string) {
	e.name = name
}

func (e *ExtensionMatcher) Match() chan core.Result {
	ext := filepath.Ext(e.name)
	ch := make(chan core.Result)
	if len(ext) > 0 {
		fmts, ok := e.m[ext]
		if ok {
			go func() {
				for _, v := range fmts {
					ch <- core.Result(v)
				}
				close(ch)
			}()
			return ch
		}
	}
	close(ch)
	return ch
}
