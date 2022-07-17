//go:build !static

package main

import "github.com/richardlehane/siegfried"

func load(path string) (*siegfried.Siegfried, error) {
	return siegfried.Load(path)
}
