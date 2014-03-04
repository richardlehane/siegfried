package namematcher

import "github.com/richardlehane/siegfried/pkg/core"

type NameMatcher interface {
	core.Matcher
	SetName(string)
}
