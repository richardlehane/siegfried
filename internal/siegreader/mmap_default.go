//go:build !windows && !linux && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd

package siegreader

func mmapable(_ int64) bool {
	return false
}

func (m *mmap) mapFile() error {
	return nil
}

func (m *mmap) unmap() error {
	return nil
}
