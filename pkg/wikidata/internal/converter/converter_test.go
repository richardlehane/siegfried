// Copyright 2020 Ross Spencer, Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// Tests to ensure the efficacy of the converter package.

package converter

import (
	"encoding/hex"
	"encoding/json"
	"testing"
)

// signature is the structure we're using to express different information
// about these tests.
type signature struct {
	Signature    string // Signature represents the original byte sequence.
	Encoding     string // Signature represents the original encoding, e.g. Hexadecimal, ASCII, PRONOM.
	NewEncoding  string // NewEncoding represents the converted encoding of the sequence we input.
	NewSignature string // NewSignature represents the converted form of the sequence we input.
	Comment      string // Comment field to potentially replay information to the user if a test fails.
	Fail         bool   // Fail flag to enable us to know when a test is expected to pass or fail.
	Converted    bool   // Converted will tell us whether the signature should have been converted or not.
}

// TestParse provides us with a way to loop through our fixtures and
// make sure the results are as expected.
func TestParse(t *testing.T) {
	var sigs []signature
	err := json.Unmarshal([]byte(testPatterns), &sigs)
	if err != nil {
		t.Error("Failed to load fixtures:", err)
	}
	for _, sig := range sigs {
		signature, converted, encoding, err := Parse(sig.Signature, LookupEncoding(sig.Encoding))
		if sig.Converted != converted {
			t.Errorf("Signature '%s' should not have been converted", sig.Signature)
		}
		if sig.NewEncoding != "" {
			if converted != true && sig.Converted {
				t.Error("Converted flag should be set to 'true'")
			}
			newEncodingReversed := ReverseEncoding(encoding)
			if sig.NewEncoding != newEncodingReversed {
				t.Errorf("Encoding conversion didn't work got '%s' expected '%s'", newEncodingReversed, sig.NewEncoding)
			}

			if sig.NewSignature != signature {
				t.Errorf("Newly encoded signature should be '%s' not '%s'", sig.NewSignature, signature)
			}
		}
		if err != nil && sig.Fail != true {
			t.Error("Failed to parse signature:", err, sig.Signature)
		}
	}
}

// TestParseEmojiRoundTrip is a little bit of a hangover from the
// olden-days. Make sure that we can round-trip strings without a loss
// of fidelity.
func TestParseEmojiRoundTrip(t *testing.T) {
	const chessEmoji = "♕♖♗♘♙♚♛♜♝♞♟"
	val, _, _, err := Parse(chessEmoji, ASCIIEncoding)
	if err != nil {
		t.Error("Failed to parse signature:", err)
	}
	roundTrip, err := hex.DecodeString(val)
	if string(roundTrip) != chessEmoji || err != nil {
		t.Errorf("Round tripping emoji failed, expected '%s', actual: '%s' (%s)", chessEmoji, val, err)
	}
}
