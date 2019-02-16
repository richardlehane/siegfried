// +build go1.5

package webarchive

import "bufio"

func discard(r *bufio.Reader, i int) (int, error) {
	return r.Discard(i)
}
