package containermatcher

import (
	"strings"

	"github.com/richardlehane/mscfb"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

type mscfbReader struct {
	rdr   *mscfb.Reader
	entry *mscfb.DirectoryEntry
}

func newMscfb(b *siegreader.Buffer) (Reader, error) {
	m, err := mscfb.New(b.NewReader())
	if err != nil {
		return nil, err
	}
	return &mscfbReader{rdr: m}, nil
}

func (m *mscfbReader) Next() error {
	var err error
	// scan to stream or error
	for m.entry, err = m.rdr.Next(); err == nil && !m.entry.Stream; m.entry, err = m.rdr.Next() {
	}
	return err
}

func (m *mscfbReader) Name() string {
	if m.entry == nil {
		return ""
	}
	path := append(m.entry.Path, m.entry.Name)
	return strings.Join(path, "/")
}

func (m *mscfbReader) SetSource(b *siegreader.Buffer) error {
	return b.SetSource(m.rdr)
}

func (m *mscfbReader) Close() {}

func (m *mscfbReader) Quit() {
	m.rdr.Quit()
}
