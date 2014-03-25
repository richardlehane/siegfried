package namematcher

import (
	"path/filepath"

	"github.com/richardlehane/siegfried/pkg/core"
)

type ExtensionMatcher map[string][]int

func NewExtensionMatcher() *ExtensionMatcher {
	em := make(map[string][]int)
	return &em
}

func (e *ExtensionMatcher) Add(ext string, fmt int) {
	_, ok := e[ext]
	if ok {
		e[ext] = append(e[ext], fmt)
		return
	}
	e[ext] = []int{fmt}
}

func (e *ExtensionMatcher) Identify(name string) chan int {
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
