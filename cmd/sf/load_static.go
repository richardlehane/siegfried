//go:build static

package main

import (
	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/static"
)

func load(path string) (*siegfried.Siegfried, error) {
	return static.New(), nil
}
