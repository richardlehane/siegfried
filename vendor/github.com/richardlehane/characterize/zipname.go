// Copyright 2015 Richard Lehane.

// +build !appengine

package characterize

import "golang.org/x/text/encoding/charmap"

// ZipName decodes names in zip files.
// If extended or international text is detected, returns IBM437 decoded string.
// Otherwise assumes UTF8 or ASCII.
func ZipName(in string) string {
	switch Detect([]byte(in)) {
	case EXTENDED, LATIN1:
		dec := charmap.CodePage437.NewDecoder()
		ret := make([]byte, len(in))
		dec.Transform(ret, []byte(in), true)
		return string(ret)
	default:
		return in
	}
}
