// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!appengine,!go1.4 dragonfly,!go1.4

package siegreader

import "syscall"

func mmapable(sz int64) bool {
	if int64(int(sz+4095)) != sz+4095 {
		return false
	}
	return true
}

func (m *mmap) mapFile() error {
	var err error
	m.buf, err = syscall.Mmap(int(m.src.Fd()), 0, int(m.sz), syscall.PROT_READ, syscall.MAP_SHARED)
	return err
}

func (m *mmap) unmap() error {
	return syscall.Munmap(m.buf)
}
