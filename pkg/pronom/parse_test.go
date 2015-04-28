package pronom

import (
	"bytes"
	"testing"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

var bsStub1 = mappings.ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "02{2}[01:1C][01:1F]????[00:03]([41:5A][61:7A]){10}(43|4E|4C)",
}

var bsStub2 = mappings.ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "02{2}000000??[00:03]([41:5A]|[61:7A]){10}(43|4E|4C)",
}

var bsStub3 = mappings.ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "5033(20|09|0D0A|0A)",
}

var bsStub4 = mappings.ByteSequence{
	Position:  "Absolute from EOF",
	Offset:    "0",
	MaxOffset: "4",
	Hex:       "(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)(30|31|32|33|34|35|36|37|38|39|20|0A|0D)20",
}

var bsStub5 = mappings.ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "264",
	Hex:       "7E56*564552532E*322E*(4C4153|43574C53)20(4C4F47204153434949205354414E44415244|4C6F67204153434949205374616E64617264|6C6F67204153434949205374616E64617264|4C4153){1-3}56(455253494F4E|657273696F6E)20322E[30:31]*7E57*7E43{5-*}7E41",
}

var bsStub6 = mappings.ByteSequence{
	Position:  "Absolute from BOF",
	Offset:    "0",
	MaxOffset: "",
	Hex:       "2F322E[30:33](0D0A|0A)2850726F6A6563742E31(0D0A|0A)094E616D653A0922",
}

var csStub = mappings.SubSequence{
	Position:        1,
	SubSeqMinOffset: "0",
	SubSeqMaxOffset: "128",
	Sequence:        "'office:document-content'",
}

var csStub1 = mappings.SubSequence{
	Position:        2,
	SubSeqMinOffset: "0",
	SubSeqMaxOffset: "",
	Sequence:        "'office:version=' [22 27] '1.0' [22 27]",
}

var csStub3 = mappings.SubSequence{
	Position:        1,
	SubSeqMinOffset: "40",
	SubSeqMaxOffset: "1064",
	Sequence:        "0F 00 00 00 'MSProject.MPP9' 00",
}

var ciStub = mappings.InternalSignature{
	ByteSequences: []mappings.ByteSeq{mappings.ByteSeq{
		SubSequences: []mappings.SubSequence{csStub, csStub1}}},
}

var ciStub1 = mappings.InternalSignature{
	ByteSequences: []mappings.ByteSeq{mappings.ByteSeq{
		SubSequences: []mappings.SubSequence{csStub3}}},
}

var sStub1 = mappings.Signature{[]mappings.ByteSequence{bsStub1}}

var sStub2 = mappings.Signature{[]mappings.ByteSequence{bsStub2}}

var sStub3 = mappings.Signature{[]mappings.ByteSequence{bsStub3, bsStub4}}

var sStub4 = mappings.Signature{[]mappings.ByteSequence{bsStub5}}

var sStub5 = mappings.Signature{[]mappings.ByteSequence{bsStub6}}

var rStub1 = &mappings.Report{Signatures: []mappings.Signature{sStub1, sStub2}, Identifiers: []mappings.FormatIdentifier{mappings.FormatIdentifier{Typ: "PUID", Id: "x-fmt/8"}}}

var rStub2 = &mappings.Report{Signatures: []mappings.Signature{sStub3}, Identifiers: []mappings.FormatIdentifier{mappings.FormatIdentifier{Typ: "PUID", Id: "x-fmt/178"}}}

var rStub3 = &mappings.Report{Signatures: []mappings.Signature{sStub4}, Identifiers: []mappings.FormatIdentifier{mappings.FormatIdentifier{Typ: "PUID", Id: "fmt/390"}}}

var rStub4 = &mappings.Report{Signatures: []mappings.Signature{sStub5}, Identifiers: []mappings.FormatIdentifier{mappings.FormatIdentifier{Typ: "PUID", Id: "x-fmt/317"}}}

func TestProcessText(t *testing.T) {
	byts := processText(csStub3.Sequence)
	if !bytes.Equal(byts, []byte{15, 0, 0, 0, 77, 83, 80, 114, 111, 106, 101, 99, 116, 46, 77, 80, 80, 57, 0}) {
		t.Fatalf("Got %v", byts)
	}
}

func TestProcessGroup(t *testing.T) {
	// try PRONOM form
	l := lexPRONOM("test", "(FF|10[!00:10])")
	<-l.items // discard group entry
	pat, err := processGroup(l)
	if err != nil {
		t.Fatal(err)
	}
	expect := patterns.Choice{
		patterns.Sequence([]byte{255}),
		patterns.List{
			patterns.Sequence([]byte{16}),
			patterns.Not{Range{[]byte{0}, []byte{16}}},
		},
	}
	if !pat.Equals(expect) {
		t.Errorf("expecting %v, got %v", expect, pat)
	}
	// try container form
	l = lexPRONOM("test2", "[10 'cats']")
	<-l.items
	pat, err = processGroup(l)
	if err != nil {
		t.Fatal(err)
	}
	expect = patterns.Choice{
		patterns.Sequence([]byte{16}),
		patterns.Sequence([]byte("cats")),
	}
	if !pat.Equals(expect) {
		t.Errorf("expecting %v, got %v", expect, pat)
	}
	// try simple
	l = lexPRONOM("test3", "[00:10]")
	<-l.items
	pat, err = processGroup(l)
	if err != nil {
		t.Fatal(err)
	}
	rng := Range{[]byte{0}, []byte{16}}
	if !pat.Equals(rng) {
		t.Errorf("expecting %v, got %v", expect, rng)
	}
}

func TestParseHex(t *testing.T) {
	ts, _, _, err := process("x-fmt/8", bsStub1.Hex, false)
	if err != nil {
		t.Error("Parse items: Error", err)
	}
	if len(ts) != 6 {
		t.Error("Parse items: Expecting 6 patterns, got", len(ts))
	}
	tok := ts[5]
	if tok.Min() != 10 || tok.Max() != 10 {
		t.Error("Parse items: Expecting 10,10, got", tok.Min(), tok.Max())
	}
	tok = ts[3]
	if tok.Min() != 2 || tok.Max() != 2 {
		t.Error("Parse items: Expecting 2,2, got", tok.Min(), tok.Max())
	}
	if !tok.Pat().Equals(Range{[]byte{0}, []byte{3}}) {
		t.Error("Parse items: Expecting [00:03], got", tok.Pat())
	}
	ts, _, _, err = process("fmt/390", bsStub5.Hex, false)
	tok = ts[12]
	if tok.Min() != 5 || tok.Max() != -1 {
		t.Error("Parse items: Expecting 5-0, got", tok.Min(), tok.Max())
	}
	if !tok.Pat().Equals(patterns.Sequence(processText("7E41"))) {
		t.Error("Parse items: Expecting 7E41, got", tok.Pat())
	}
	ts, _, _, err = process("x-fmt/317", bsStub6.Hex, false)
	seqs := ts[2].Pat().Sequences()
	if !seqs[0].Equals(patterns.Sequence(processText("0D0A"))) {
		t.Error("Parse items: Expecting [13 10], got", []byte(seqs[0]))
	}

}

func TestParseReports(t *testing.T) {
	r := &reports{[]string{"test1", "test2", "test3", "test4"}, []*mappings.Report{rStub1, rStub2, rStub3, rStub4}, nil}
	_, _, err := r.signatures()
	if err != nil {
		t.Error(err)
	}
}

func TestParseContainer(t *testing.T) {
	sig, err := processDROID("fmt/123", ciStub.ByteSequences)
	if err != nil {
		t.Error(err)
	}
	if len(sig) != 5 {
		t.Error("Expecting 5 patterns! Got ", sig)
	}
	sig, err = processDROID("fmt/123", ciStub1.ByteSequences)
	if err != nil {
		t.Error(err)
	}
	if min, _ := sig[0].Length(); min != 19 {
		t.Error("Expecting a sequence with a length of 19! Got ", sig)
	}
}
