package namematcher

import (
	"path/filepath"
	"strings"
)

type ExtensionMatcher map[string][]int

func NewExtensionMatcher() ExtensionMatcher {
	return make(ExtensionMatcher)
}

func (e ExtensionMatcher) Add(ext string, fmt int) {
	_, ok := e[ext]
	if ok {
		e[ext] = append(e[ext], fmt)
		return
	}
	e[ext] = []int{fmt}
}

func (e ExtensionMatcher) Identify(name string) []int {
	ext := filepath.Ext(name)
	if len(ext) > 0 {
		fmts, ok := e[strings.TrimPrefix(ext, ".")]
		if ok {
			return fmts
		}
	}
	return []int{}
}
