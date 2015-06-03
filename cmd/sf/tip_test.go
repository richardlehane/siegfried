package main

import (
	"bytes"
	"testing"
)

func TestTip(t *testing.T) {
	expect := "fmt/669"
	err := setup()
	if err != nil {
		t.Error(err)
	}
	buf := bytes.NewReader([]byte{0x00, 0x4d, 0x52, 0x4d, 0x00})
	c, err := s.Identify("test.mrw", buf)
	for i := range c {
		if i.String() != expect {
			t.Errorf("First buffer: expecting %s, got %s", expect, i)
		}
	}
	buf = bytes.NewReader([]byte{0x00, 0x4d, 0x52, 0x4d, 0x00})
	c, err = s.Identify("test.mrw", buf)
	for i := range c {
		if i.String() != expect {
			t.Errorf("Second buffer: expecting %s, got %s", expect, i)
		}
	}
	buf = bytes.NewReader([]byte{0x00, 0x4d, 0x52, 0x4d, 0x00})
	c, err = s.Identify("test.mrw", buf)
	for i := range c {
		if i.String() != expect {
			t.Errorf("Third buffer: expecting %s, got %s", expect, i)
		}
	}
}
