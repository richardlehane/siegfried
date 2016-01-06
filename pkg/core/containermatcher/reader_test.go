package containermatcher

import (
	"bytes"
	"io"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type node struct {
	name   string
	stream []byte
}

type testReader struct {
	nodes []*node
	idx   int
}

func (tr *testReader) Next() error {
	tr.idx++
	if tr.idx >= len(tr.nodes)-1 {
		return io.EOF
	}
	return nil
}

func (tr *testReader) Name() string {
	return tr.nodes[tr.idx].name
}

func (tr *testReader) SetSource(b *siegreader.Buffers) (*siegreader.Buffer, error) {
	return b.Get(bytes.NewReader(tr.nodes[tr.idx].stream))
}

func (tr *testReader) Close() {}

func (tr *testReader) Quit() {}

var ns []*node = []*node{
	&node{
		"one",
		[]byte("test12345678910YNESSjunktestyjunktestytest12345678910111223"),
	},
	&node{
		"two",
		[]byte("test12345678910YNESSjunktestyjunktestytest12345678910111223"),
	},
	&node{
		"three",
		[]byte("test12345678910YNESSjunktestyjunktestytest12345678910111223"),
	},
}

var tr *testReader = &testReader{nodes: ns}

func newTestReader(buf *siegreader.Buffer) (Reader, error) {
	tr.idx = -1
	return tr, nil
}

func TestReader(t *testing.T) {
	tr.idx = -1
	err := tr.Next()
	if err != nil {
		t.Error(err)
	}
	_ = tr.Next()
	err = tr.Next()
	if err != io.EOF {
		t.Error("expecting EOF")
	}
}
