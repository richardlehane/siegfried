package webarchive

import (
	"os"
	"testing"
)

func TestDecodePayload(t *testing.T) {
	f, _ := os.Open("examples/decode.warc")
	defer f.Close()
	rdr, err := NewWARCReader(f)
	if err != nil {
		t.Fatal("failure loading example: " + err.Error())
	}
	rec, err := rdr.NextPayload()
	if err != nil {
		t.Fatal(err)
	}
	dec := DecodePayload(rec)
	buf := make([]byte, 4)
	if i, err := dec.Read(buf); err != nil || i != 4 {
		t.Fatal("failure reading decode.warc")
	}
	if string(buf) != "\n<!D" {
		t.Fatalf("expecting ' <!D' got %s", buf)
	}
}

func TestDecodePayloadT(t *testing.T) {
	f, _ := os.Open("examples/decode.warc")
	defer f.Close()
	rdr, err := NewWARCReader(f)
	if err != nil {
		t.Fatal("failure loading example: " + err.Error())
	}
	rec, err := rdr.NextPayload()
	if err != nil {
		t.Fatal(err)
	}
	dec := DecodePayloadT(rec)
	buf := make([]byte, 4)
	if i, err := dec.Read(buf); err != nil || i != 4 {
		t.Fatal("failure reading decode.warc")
	}
	if string(buf) == "\n<!D" {
		t.Fatalf("expecting gibberish got %s", buf)
	}
}
