// Tests the siegfried output encoding and makes sure that the writers
// encode special characters correctly, e.g. so that if the caller
// outputs JSON then it is valid JSON.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/pronom"
	"github.com/richardlehane/siegfried/pkg/writer"
)

var controlCharacters = []string{"\u0000", "\u0001", "\u0002", "\u0003",
	"\u0004", "\u0005", "\u0006", "\u0007", "\u0008", "\u0009", "\u000A",
	"\u000B", "\u000C", "\u000D", "\u000E", "\u000F", "\u0010", "\u0011",
	"\u0012", "\u0013", "\u0014", "\u0015", "\u0016", "\u0017", "\u0018",
	"\u0019",
}
var nonControlCharacters = []string{"\u0020", "\u1F5A4", "\u265B", "\u1F0A1",
	"\u262F",
}

var jsonSiegfried *siegfried.Siegfried

func setupOutputEncodingTests(pronomx bool) error {
	jsonSiegfried = siegfried.New()
	config.SetHome(*testhome)
	opts := []config.Option{}
	identifier, err := pronom.New(opts...)
	if err != nil {
		return err
	}
	jsonSiegfried.Add(identifier)
	return nil
}

// TestNonControlCharacters tests control characters that are valid but
// need special treatment from the writer and makes sure that they
// create invalid JSON.
//
// When the tests pass, this test will be reworked to test for valid
// JSON instead of a True-Negative.
//
func TestControlCharacters(t *testing.T) {
	err := setupOutputEncodingTests(false)
	if err != nil {
		t.Error(err)
	}

	// Run some data through the identifier so that we have something
	// useful on the other side.
	expected := "UNKNOWN"
	badfoodBuffer := bytes.NewReader([]byte{0xba, 0xdf, 0x00, 0xd0})
	res, _ := jsonSiegfried.Identify(badfoodBuffer, "", "")
	for _, val := range res {
		if val.String() != expected {
			t.Errorf("Expecting %s, got %s", expected, val)
		}
	}

	// Loop through the control characters to make sure the JSON output
	// is invalid.
	for _, val := range controlCharacters {
		path := fmt.Sprintf("path/%sto/file", val)
		var w writer.Writer
		buf := new(bytes.Buffer)
		w = writer.JSON(buf)
		w.Head(path, time.Now(), time.Now(), [3]int{0, 0, 0},
			jsonSiegfried.Identifiers(),
			jsonSiegfried.Fields(),
			"md5",
		)
		w.File("testName", 10, "testMod", []byte("d41d8c"), nil, res)
		w.Tail()
		if json.Valid([]byte(buf.String())) {
			t.Errorf("Expecting control characters '0x%x' to create invalid JSON: %s",
				val, buf.String(),
			)
		}
	}
}

// TestNonControlCharacters tests valid characters and simply makes sure
// that the JSON output is correct.
func TestNonControlCharacters(t *testing.T) {
	err := setupOutputEncodingTests(false)
	if err != nil {
		t.Error(err)
	}

	// Run some data through the identifier so that we have something
	// useful on the other side.
	expected := "UNKNOWN"
	badfoodBuffer := bytes.NewReader([]byte{0xba, 0xdf, 0x00, 0xd0})
	res, _ := jsonSiegfried.Identify(badfoodBuffer, "", "")
	for _, val := range res {
		if val.String() != expected {
			t.Errorf("Expecting %s, got %s", expected, val)
		}
	}

	// Loop through the valid non-control characters to make sure the
	// JSON is valid regardless.
	for _, val := range nonControlCharacters {
		path := fmt.Sprintf("path/%sto/file", val)
		var w writer.Writer
		buf := new(bytes.Buffer)
		w = writer.JSON(buf)
		w.Head(path, time.Now(), time.Now(), [3]int{0, 0, 0},
			jsonSiegfried.Identifiers(),
			jsonSiegfried.Fields(),
			"md5",
		)
		w.File("testName", 10, "testMod", []byte("d41d8c"), nil, res)
		w.Tail()
		if !json.Valid([]byte(buf.String())) {
			t.Errorf("Expecting valid character '0x%x' to create valid JSON: %s",
				val, buf.String(),
			)
		}
	}
}
