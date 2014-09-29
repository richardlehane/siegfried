package mappings

import "encoding/xml"

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
	Path      string
	Signature InternalSignature `xml:"BinarySignatures>InternalSignatureCollection>InternalSignature"`
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
	Id   int    `xml:"signatureId,attr"`
	Puid string `xml:",attr"`
}

type TriggerPuid struct {
	ContainerType string `xml:",attr"`
	Puid          string `xml:",attr"`
}
