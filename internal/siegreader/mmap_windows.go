// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

package siegreader

import (
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

func mmapable(sz int64) bool {
	if int64(int(sz+4095)) != sz+4095 {
		return false
	}
	return true
}

func (m *mmap) mapFile() error {
	h, err := syscall.CreateFileMapping(syscall.Handle(m.src.Fd()), nil, syscall.PAGE_READONLY, uint32(m.sz>>32), uint32(m.sz), nil)
	if err != nil {
		return err
	}
	m.handle = uintptr(h) // for later unmapping
	addr, err := syscall.MapViewOfFile(h, syscall.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		return err
	}
	m.buf = []byte{}
	slcHead := (*reflect.SliceHeader)(unsafe.Pointer(&m.buf))
	slcHead.Data = addr
	slcHead.Len = int(m.sz)
	slcHead.Cap = int(m.sz)
	return nil
}

func (m *mmap) unmap() error {
	slcHead := (*reflect.SliceHeader)(unsafe.Pointer(&m.buf))
	err := syscall.UnmapViewOfFile(slcHead.Data)
	if err != nil {
		return err
	}
	return os.NewSyscallError("CloseHandle", syscall.CloseHandle(syscall.Handle(m.handle)))
}
