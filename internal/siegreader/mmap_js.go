// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js

package siegreader

func mmapable(sz int64) bool {
	return false
}

func (m *mmap) mapFile() error {
	var err error
	return err
}

func (m *mmap) unmap() error {
	var err error
	return err
}
