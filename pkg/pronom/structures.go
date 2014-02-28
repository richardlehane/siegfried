// This file contains struct mappings to unmarshal three different PRONOM XML formats: the signature file format, the report format, and the container format
package pronom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// Droid Signature File

type Droid struct {
	XMLName     xml.Name     `xml:"FFSignatureFile"`
	Version     int          `xml:",attr"`
	FileFormats []FileFormat `xml:"FileFormatCollection>FileFormat"`
}

func (d Droid) String() string {
	buf := new(bytes.Buffer)
	for _, v := range d.FileFormats {
		fmt.Fprintln(buf, v)
	}
	return buf.String()
}

type FileFormat struct {
	ID         int      `xml:",attr"`
	Puid       string   `xml:"PUID,attr"`
	Name       string   `xml:",attr"`
	Version    string   `xml:",attr"`
	MIMEType   string   `xml:",attr"`
	Extensions []string `xml:"Extension"`
	Priorities []int    `xml:"HasPriorityOverFileFormatID"`
	*Report
}

func (f FileFormat) String() string {
	null := func(s string) string {
		if strings.TrimSpace(s) == "" {
			return "NULL"
		}
		return s
	}
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "Puid: %s; Name: %s; Version: %s; Ext(s): %s\n", f.Puid, f.Name, f.Version, strings.Join(f.Extensions, ", "))
	fmt.Fprintln(buf, f.Description)
	for _, v := range f.Signatures {
		fmt.Fprint(buf, "Signature\n")
		for _, v1 := range v.ByteSequences {
			fmt.Fprintf(buf, "Position: %s, Offset: %s, MaxOffset: %s, IndirectLoc: %s, IndirectLen: %s, Endianness: %s\n",
				null(v1.Position), null(v1.Offset), null(v1.MaxOffset), null(v1.IndirectLoc), null(v1.IndirectLen), null(v1.Endianness))
			fmt.Fprintln(buf, v1.Hex)
		}
	}
	return buf.String()
}

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

// Container

type Container struct {
	XMLName             xml.Name             `xml:"ContainerSignatureMapping"`
	ContainerSignatures []ContainerSignature `xml:"ContainerSignatures>ContainerSignature"`
	FormatMappings      []FormatMapping      `xml:"FileFormatMappings>FileFormatMapping"`
	TriggerPuids        []TriggerPuid        `xml:"TriggerPuids>TriggerPuid"`
}

type ContainerSignature struct {
	Id            int    `xml:",attr"`
	ContainerType string `xml:",attr"`
	Description   string
	Files         []File `xml:"Files>File"`
}

type File struct {
	Required  bool `xml:",attr"`
	Path      string
	Signature InternalSignature `xml:"BinarySignatures>InternalSignatureCollection"`
}

type InternalSignature struct {
	Id            int       `xml:"ID,attr"`
	Specificity   string    `xml:",attr"`
	ByteSequences []ByteSeq `xml:"ByteSequence"`
}

type ByteSeq struct {
	Reference    string        `xml:"Reference,attr"`
	SubSequences []SubSequence `xml:"SubSequence"`
}

type SubSequence struct {
	Position        int    `xml:",attr"`
	SubSeqMinOffset string `xml:",attr"` // and empty int values are unmarshalled to 0
	SubSeqMaxOffset string `xml:",attr"` // uses string rather than int because value might be empty
	Sequence        string
}

type FormatMapping struct {
	Id   string `xml:"signatureId,attr"`
	Puid string `xml:",attr"`
}

type TriggerPuid struct {
	ContainerType string `xml:",attr"`
	Puid          string `xml:",attr"`
}
