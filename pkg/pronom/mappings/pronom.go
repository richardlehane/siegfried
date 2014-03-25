package mappings

import (
	"encoding/xml"
	"strings"
)

// PRONOM Report

type Report struct {
	XMLName     xml.Name    `xml:"PRONOM-Report"`
	Description string      `xml:"report_format_detail>FileFormat>FormatDescription"`
	Signatures  []Signature `xml:"report_format_detail>FileFormat>InternalSignature"`
}

type Signature struct {
	ByteSequences []ByteSequence `xml:"ByteSequence"`
}

func (s Signature) String() string {
	var p string
	for i, v := range s.ByteSequences {
		if i > 0 {
			p += "\n"
		}
		p += v.String()
	}
	return p
}

type ByteSequence struct {
	Position    string `xml:"PositionType"`
	Offset      string
	MaxOffset   string
	IndirectLoc string `xml:"IndirectOffsetLocation"`
	IndirectLen string `xml:"IndirectOffsetLength"`
	Endianness  string
	Hex         string `xml:"ByteSequenceValue"`
}

func trim(label, s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return label + ":" + s + " "
}

func (bs ByteSequence) String() string {
	return trim("Pos", bs.Position) + trim("Off", bs.Offset) + trim("Max", bs.MaxOffset) + trim("Hex", bs.Hex)
}
