// +build darwin,!go1.4 freebsd,!go1.4 netbsd,!go1.4 openbsd,!go1.4

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
	m.buf, err = syscall.Mmap(int(m.src.Fd()), 0, int(m.sz), 1, 1)
	return err
}

func (m *mmap) unmap() error {
	return syscall.Munmap(m.buf)
}
