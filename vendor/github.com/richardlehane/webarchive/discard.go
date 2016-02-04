// +build !go1.5

package webarchive

import (
	"bufio"
)

var discardBuf []byte

func discard(r *bufio.Reader, i int) (int, error) {
	if len(discardBuf) < i {
		discardBuf = make([]byte, i)
	}
	l, err := fullRead(r, discardBuf[:i])
	if l != i {
		return l, ErrDiscard
	}
	return l, err
}
