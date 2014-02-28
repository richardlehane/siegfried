package pronom

import (
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
)

var bsStub1 = ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "02{2}[01:1C][01:1F]????[00:03]([41:5A][61:7A]){10}(43|4E|4C)",
}

var bsStub2 = ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "02{2}000000??[00:03]([41:5A]|[61:7A]){10}(43|4E|4C)",
}

var bsStub3 = ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "5033(20|09|0D0A|0A)",
}

var bsStub4 = ByteSequence{
	Position:  "Absolute from EOF",
	Offset:    "0",
	MaxOffset: "4",
	Hex:       "(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)20",
}

var bsStub5 = ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "264",
	Hex:       "7E56*564552532E*322E*(4C4153|43574C53)20(4C4F47204153434949205354414E44415244|4C6F67204153434949205374616E64617264|6C6F67204153434949205374616E64617264|4C4153){1-3}56(455253494F4E|657273696F6E)20322E[30:31]*7E57*7E43{5-*}7E41",
}

var bsStub6 = ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "2F322E[30:33](0D0A|0A)2850726F6A6563742E31(0D0A|0A)094E616D653A0922",
}

var sStub1 = Signature{[]ByteSequence{bsStub1}}

var sStub2 = Signature{[]ByteSequence{bsStub2}}

var sStub3 = Signature{[]ByteSequence{bsStub3, bsStub4}}

var sStub4 = Signature{[]ByteSequence{bsStub5}}

var sStub5 = Signature{[]ByteSequence{bsStub6}}

var rStub1 = &Report{Signatures: []Signature{sStub1, sStub2}}

var rStub2 = &Report{Signatures: []Signature{sStub3}}

var rStub3 = &Report{Signatures: []Signature{sStub4}}

var rStub4 = &Report{Signatures: []Signature{sStub5}}

var fStub1 = FileFormat{
	Puid:   "x-fmt/8",
	Report: rStub1,
}

var fStub2 = FileFormat{
	Puid:   "x-fmt/178",
	Report: rStub2,
}

var fStub3 = FileFormat{
	Puid:   "fmt/390",
	Report: rStub3,
}

var fStub4 = FileFormat{
	Puid:   "x-fmt/317",
	Report: rStub4,
}

var pStub = &pronom{
	droid: &Droid{FileFormats: []FileFormat{fStub1, fStub2, fStub3, fStub4}},
}

func TestParseHex(t *testing.T) {
	ts, _, _, err := parseHex("x-fmt/8", bsStub1.Hex)
	if err != nil {
		t.Error("Parse items: Error", err)
	}
	if len(ts) != 6 {
		t.Error("Parse items: Expecting 6 patterns, got", len(ts))
	}
	tok := ts[5]
	if tok.min != 10 || tok.max != 10 {
		t.Error("Parse items: Expecting 10,10, got", tok.min, tok.max)
	}
	tok = ts[3]
	if tok.min != 2 || tok.max != 2 {
		t.Error("Parse items: Expecting 2,2, got", tok.min, tok.max)
	}
	if !tok.pat.Equals(Range{[]byte{0}, []byte{3}}) {
		t.Error("Parse items: Expecting [00:03], got", tok.pat)
	}
	ts, _, _, err = parseHex("fmt/390", bsStub5.Hex)
	tok = ts[12]
	if tok.min != 5 || tok.max != -1 {
		t.Error("Parse items: Expecting 5-0, got", tok.min, tok.max)
	}
	if !tok.pat.Equals(bytematcher.Sequence(decodeHex("7E41"))) {
		t.Error("Parse items: Expecting 7E41, got", tok.pat)
	}
	ts, _, _, err = parseHex("x-fmt/317", bsStub6.Hex)
	seqs := ts[2].pat.Sequences()
	if !seqs[0].Equals(bytematcher.Sequence(decodeHex("0D0A"))) {
		t.Error("Parse items: Expecting [13 10], got", []byte(seqs[0]))
	}

}

func TestParse(t *testing.T) {
	sigs, err := pStub.Parse()
	if err != nil {
		t.Error(err)
	}
	if len(sigs) < 2 {
		t.Error("Expecting more patterns than that! Got ", len(sigs))
	}
}
