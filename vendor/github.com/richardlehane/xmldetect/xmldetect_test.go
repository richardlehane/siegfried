package xmldetect

import (
	"io"
	"strings"
	"testing"
)

const (
	nofail = 1 << iota
	errfail
	decfail
	rootfail
	nsfail
	closefail
)

var tests = []struct {
	fails       int // which values in this test case are *deliberate* fails?
	err         error
	declaration bool
	root        string
	ns          string
	content     string // sample xml
}{
	{
		nofail,
		nil,
		true,
		"hello",
		"http://www.adele.com",
		`<?xml version="1.0" encoding="UTF-8" ?>
		<!-- comment --->
		<hello xmlns="http://www.adele.com">
			<from>
				the other
				</side>
			</from>
		</hello>`,
	},
	{
		nofail,
		nil,
		false,
		"doc",
		"",
		`<!DOCTYPE doc [
		<!ELEMENT doc (#PCDATA)>
		<!ENTITY e "<![CDATA[&foo;]]]]]]]]>">
		]>
		<doc>&e;</doc>

		<?bogus test="true"?>
		<!-- an end comment, nasty!-->`},
	{
		decfail | rootfail | nsfail | closefail,
		nil,
		true,
		"goodbye",
		"http://www.pink.com",
		`<!-- comment -->
		<hello xmlns="http://www.adele.com">
			<from>
				the other
				</side>
			</from>
		</hello>`,
	},
}

func TestRoot(t *testing.T) {
	for idx, v := range tests {
		d, r, ns, err := Root(strings.NewReader(v.content))
		// error
		if err != v.err {
			if v.fails&errfail != errfail {
				t.Errorf("testcase %d: bad err, got %v, expected %v", idx, err, v.err)
			}
		} else if v.fails&errfail == errfail {
			t.Errorf("testcase %d: bad err, got %v, expected %v to fail", idx, err, v.err)
		}
		// declaration
		if d != v.declaration {
			if v.fails&decfail != decfail {
				t.Errorf("testcase %d: bad declaration, got %v, expected %v", idx, d, v.declaration)
			}
		} else if v.fails&decfail == decfail {
			t.Errorf("testcase %d: bad declaration, got %v, expected %v to fail", idx, d, v.declaration)
		}
		// opening tag
		if r != v.root {
			if v.fails&rootfail != rootfail {
				t.Errorf("testcase %d: bad root, got %s, expected %s", idx, r, v.root)
			}
		} else if v.fails&rootfail == rootfail {
			t.Errorf("testcase %d: bad root, got %s, expected %s to fail", idx, r, v.root)
		}
		// namespace
		if ns != v.ns {
			if v.fails&nsfail != nsfail {
				t.Errorf("testcase %d: bad namespace, got %s, expected %s", idx, ns, v.ns)
			}
		} else if v.fails&nsfail == nsfail {
			t.Errorf("testcase %d: bad namespace, got %s, expected %s to fail", idx, ns, v.ns)
		}
	}
}

type reverseReader struct {
	str string
	idx int
}

func (rr *reverseReader) ReadByte() (c byte, err error) {
	if rr.idx > len(rr.str)-1 {
		return 0, io.EOF
	}
	rr.idx++
	return rr.str[len(rr.str)-rr.idx], nil
}

func TestClosing(t *testing.T) {
	for idx, v := range tests {
		tag, err := Closing(&reverseReader{str: v.content})
		// error
		if err != v.err {
			if v.fails&errfail != errfail {
				t.Errorf("testcase %d: bad err, got %v, expected %v", idx, err, v.err)
			}
		} else if v.fails&errfail == errfail {
			t.Errorf("testcase %d: bad err, got %v, expected %v to fail", idx, err, v.err)
		}
		// closing tag
		if tag != v.root {
			if v.fails&closefail != closefail {
				t.Errorf("testcase %d: bad closing tag, got %s, expected %s", idx, tag, v.root)
			}
		} else if v.fails&closefail == closefail {
			t.Errorf("testcase %d: bad closing tag, got %s, expected %s to fail", idx, tag, v.root)
		}
	}
}
