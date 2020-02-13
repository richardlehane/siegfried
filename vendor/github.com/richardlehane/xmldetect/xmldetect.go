// Copyright 2015 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xmldetect

import (
	"errors"
	"io"
)

var ErrInvalid = errors.New("invalid XML")

// Root inspects the beginning of an XML document and returns:
// - whether it has a <?xml... declaration
// - the opening tag name
// - the root default namespace (xmlns)
// - an error if it isn't valid XML.
//
// Example:
//   var ex = `
//      <?xml version="1.0"?>
//      <!DOCTYPE doc [
//		<!ELEMENT doc (#PCDATA)>
//		<!ENTITY e "<![CDATA[&foo;]]]]]]]]>">
//		]>
//      <!-- ignore me -->
//		<doc xmlns="ex">&e;</doc>`
//   dec, tag, ns, err := Root(strings.NewReader(ex))
//   if err == nil && dec && tag == "doc" && ns == "ex" {
//	   fmt.Println("root tag is 'doc' with namespace 'ex'")
//   }
func Root(in io.ByteReader) (declaration bool, tag, ns string, err error) {
	for c, e := in.ReadByte(); e == nil; c, e = in.ReadByte() {
		switch c {
		default:
			return false, "", "", ErrInvalid
		// beginning of a token
		case '<':
			d, t, n, e := eatToken(in)
			if e != nil {
				return false, "", "", e
			}
			if d {
				declaration = true
			}
			if t != "" {
				tag, ns = t, n
				return
			}
		// ignore whitespace
		case ' ', '\r', '\n', '\t':
		}
	}
	return false, "", "", ErrInvalid
}

// Closing inspects the end of an XML document and returns the closing tag.
// The supplied byte reader should read bytes from the XML document in reverse,
// e.g. see the reverseReader in the example below.
//
// Example:
//  type reverseReader struct {
//	  str string
//	  idx int
//  }
//  func (rr *reverseReader) ReadByte() (c byte, err error) {
//	  if rr.idx > len(rr.str)-1 {
//		return 0, io.EOF
//	  }
//	  rr.idx++
//	  return rr.str[len(rr.str)-rr.idx], nil
//   }
//   var ex = `
//      <?xml version="1.0"?>
//      <!DOCTYPE doc [
//		<!ELEMENT doc (#PCDATA)>
//		<!ENTITY e "<![CDATA[&foo;]]]]]]]]>">
//		]>
//		<doc>&e;</doc>
//
//		<?bogus test="true"?>
//		<!-- an end comment, nasty!-->`
//   tag, err := Closing(&reverseReader{str: ex})
//   if err == nil && tag == "doc" {
//	   fmt.Println("closing tag is 'doc'")
//   }
func Closing(in io.ByteReader) (tag string, err error) {
	for c, e := in.ReadByte(); e == nil; c, e = in.ReadByte() {
		switch c {
		default:
			return "", ErrInvalid
		case '>':
			t, e := eatBackToken(in)
			if e != nil {
				return "", e
			}
			if t != "" {
				return t, nil
			}
		// ignore whitespace
		case ' ', '\r', '\n', '\t':
		}
	}
	return "", ErrInvalid
}

var (
	piClose      = []byte("?>")
	commentClose = []byte("-->")
	commentStart = []byte("--!<")
	piStart      = []byte("?<")
)

// must eat will consume the ByteReader until the expected bytes are reached.
// Returns true if expected bytes are reached.
func mustEat(in io.ByteReader, expect []byte) bool {
	var idx int
	for {
		c, err := in.ReadByte()
		if err != nil {
			return false
		}
		// happy case
		if c == expect[idx] {
			if idx == len(expect)-1 {
				return true
			}
			idx++
			continue
		}
		// no match yet
		if idx == 0 {
			continue
		}
		// backtrack
		for i := idx; i > 0; i-- {
			if c == expect[i-1] {
				subseq := true
				for n, b := range expect[:i-1] {
					if b != expect[idx-i+n] {
						subseq = false
						break
					}
				}
				if subseq {
					idx = i
					break
				}
			}
		}
	}
}

// eats a single token returning:
// a bool, if it is an xml processing instruction
// a tag and optional default namespace, if an element
// and an error
func eatToken(in io.ByteReader) (bool, string, string, error) {
	c, err := in.ReadByte()
	if err != nil {
		return false, "", "", ErrInvalid
	}
	switch c {
	default:
		valid, tag, ns := eatElement(in, c)
		if valid {
			return false, tag, ns, nil
		}
	case '!':
		// can be a comment or a DOCTYPE
		if eatComment(in) {
			return false, "", "", nil
		}
	case '?':
		valid, decl := eatPI(in)
		if valid {
			return decl, "", "", nil
		}
	case '<', '>', ' ', '\r', '\n', '\t':
	}
	return false, "", "", ErrInvalid
}

// eats an element returning:
// a bool to signfify if a valid element
// the tag name
// the default (xmlns=) namespace.
func eatElement(in io.ByteReader, c byte) (bool, string, string) {
	buf := make([]byte, 32)
	var (
		err             error
		tag, ns, prefix string
		idx             int
	)
	for ; err == nil; c, err = in.ReadByte() {
		switch c {
		case '>', ' ', '\r', '\n', '\t':
			if tag == "" {
				tag = string(buf[:idx])
			} else if ns == "" {
				ns = extractNS(prefix, buf[:idx])
			}
			if c == '>' {
				return true, tag, ns
			}
			idx = 0
		default:
			if c == ':' && tag == "" {
				prefix = string(buf[:idx])
				idx = 0
			} else {
				if idx >= len(buf) {
					cp := make([]byte, len(buf)*2)
					copy(cp, buf)
					buf = cp
				}
				buf[idx] = c
				idx++
			}
		}
	}
	return false, "", ""
}

func extractNS(prefix string, buf []byte) string {
	var (
		inQuote    byte
		start, end int
	)
	if prefix == "" {
		start = 6
		if len(buf) > 8 &&
			(buf[0] == 'x' || buf[0] == 'X') &&
			(buf[1] == 'm' || buf[1] == 'M') &&
			(buf[2] == 'l' || buf[2] == 'L') &&
			(buf[3] == 'n' || buf[3] == 'N') &&
			(buf[4] == 's' || buf[4] == 'S') &&
			buf[5] == '=' {
		} else {
			return ""
		}
	} else {
		start = len(prefix) + 1
		if len(buf) > len(prefix)+1 {
			for i := 0; i < len(prefix); i++ {
				if buf[i] != prefix[i] {
					return ""
				}
			}
			if buf[len(prefix)] != '=' {
				return ""
			}
		} else {
			return ""
		}
	}
	for i, c := range buf[start:] {
		if inQuote == 0 {
			// test for " or '
			if c == 0x27 || c == 0x22 {
				inQuote = c
				end = start + i
				start += i + 1
			}
		} else if c == inQuote {
			return string(buf[start : end+i])
		}
	}
	return ""
}

// eats a processing instruction, returning:
// - validity
// - whether it is an xml declaration i.e. <?xml ...?>
func eatPI(in io.ByteReader) (bool, bool) {
	a, err := in.ReadByte()
	if err != nil {
		return false, false
	}
	b, err := in.ReadByte()
	if err != nil {
		return false, false
	}
	if a == '?' && b == '>' {
		return true, false
	}
	c, err := in.ReadByte()
	if err != nil {
		return false, false
	}
	if b == '?' && c == '>' {
		return true, false
	}
	for c == '?' {
		c, err = in.ReadByte()
		if err != nil {
			return false, false
		}
		if c == '>' {
			return true, false
		}
	}
	var declaration bool
	if (a == 'x' || a == 'X') &&
		(b == 'm' || b == 'M') &&
		(c == 'l' || c == 'L') {
		declaration = true
	}
	return mustEat(in, piClose), declaration
}

func eatComment(in io.ByteReader) bool {
	c, err := in.ReadByte()
	if err != nil {
		return false
	}
	// comment or DOCTYPE?
	switch c {
	case '-':
		c, _ = in.ReadByte()
		if c != '-' {
			return false
		}
		return mustEat(in, commentClose)
	case 'D':
		return eatDOCTYPE(in)
	default:
		return false
	}
}

func eatDOCTYPE(in io.ByteReader) bool {
	expect := make([]byte, 6)
	for i := range expect {
		c, err := in.ReadByte()
		if err != nil {
			return false
		}
		expect[i] = c
	}
	if string(expect) != "OCTYPE" {
		return false
	}
	var depth int
	for {
		c, err := in.ReadByte()
		if err != nil {
			return false
		}
		switch c {
		case '>':
			if depth == 0 {
				return true
			}
			depth--
		case '<':
			depth++
		}
	}
}

// eats a token in reverse, returning the closing tag name and an error.
func eatBackToken(in io.ByteReader) (string, error) {
	c, err := in.ReadByte()
	if err != nil {
		return "", ErrInvalid
	}
	switch c {
	default:
		valid, tag := eatBackElement(in, c)
		if valid {
			return tag, nil
		}
	case '-':
		c, err = in.ReadByte()
		if err == nil && c == '-' && mustEat(in, commentStart) {
			return "", nil
		}
	case '?':
		if mustEat(in, piStart) {
			return "", nil
		}
	case '<', '>':
	}
	return "", ErrInvalid
}

func eatBackElement(in io.ByteReader, c byte) (bool, string) {
	buf := make([]byte, 32)
	var (
		err error
		tag string
		idx int
	)
	var hasEnd bool
	for ; err == nil; c, err = in.ReadByte() {
		switch c {
		default:
			if idx >= len(buf) {
				cp := make([]byte, len(buf)*2)
				copy(cp, buf)
				buf = cp
			}
			buf[idx] = c
			idx++
		case '/':
			hasEnd = true
		case '<', ' ', '\r', '\n', '\t':
			if idx > 0 {
				// reverse the tag name
				for left, right := 0, len(buf[:idx])-1; left < right; left, right = left+1, right-1 {
					buf[left], buf[right] = buf[right], buf[left]
				}
				tag = string(buf[:idx])
			}
			if c == '<' {
				return hasEnd, tag
			}
			idx = 0
		}
	}
	return false, ""
}
