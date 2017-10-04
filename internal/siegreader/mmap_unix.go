// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux darwin dragonfly freebsd netbsd openbsd
// +build !appengine

package siegreader

import "golang.org/x/sys/unix"

func mmapable(sz int64) bool {
	if int64(int(sz+4095)) != sz+4095 {
		return false
	}
	return true
}

func (m *mmap) mapFile() error {
	var err error
	m.buf, err = unix.Mmap(int(m.src.Fd()), 0, int(m.sz), unix.PROT_READ, unix.MAP_SHARED)
	return err
}

func (m *mmap) unmap() error {
	return unix.Munmap(m.buf)
}
