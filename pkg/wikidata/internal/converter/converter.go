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

// Convert file-format signature sequences to something compatible with
// Siegfried's identifiers.
package converter

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Parse will take a signature and convert it into something that can
// be used. If the signature needs to be converted then the function
// will inform the caller and return the new encoding value.
func Parse(signature string, encoding int) (string, bool, int, error) {
	switch encoding {
	case HexEncoding:
		hex, err := HexParse(signature)
		if err != nil {
			return "", false, encoding, err
		}
		return hex, false, encoding, nil
	case ASCIIEncoding:
		hexEncoded := ASCIIParser(signature)
		return hexEncoded, true, HexEncoding, nil
	case PerlEncoding:
		pronomEncoded, converted := PERLParser(signature)
		if converted {
			return pronomEncoded, converted, PronomEncoding, nil
		}
		return "", false, encoding, fmt.Errorf("Not processing PERL")
	case PronomEncoding:
		return signature, false, encoding, nil
	case GUIDEncoding:
		return "",
			false,
			encoding, fmt.Errorf("Not processing GUID")
	case UnknownEncoding:
		hex, err := HexParse(signature)
		if err != nil {
			return "",
				false,
				encoding,
				fmt.Errorf("Unknown conversion to hex failed: %s", err)
		}
		return hex, true, HexEncoding, nil
	}
	return "", false, encoding, nil
}

// preprocessHex will perform some basic operations on a hexadecimal
// string to give it a greater chance of being decoded.
func preprocessHex(signature string) string {
	// Remove non-encoded spaces from HEX e.g. `AC DC` -> `ACDC`.
	signature = strings.Replace(signature, " ", "", -1)
	if strings.HasPrefix(signature, "0x") {
		// Remove 0x prefix some values have.
		signature = strings.Replace(signature, "0x", "", 1)
	}
	// Convert the hex string to upper-case to be consistent.
	signature = strings.ToUpper(signature)
	return signature
}

// HexParse will take a hexadecimal based signature and do something
// useful with it...
func HexParse(signature string) (string, error) {
	signature = preprocessHex(signature)
	// Validate the hexadecimal
	_, err := hex.DecodeString(signature)
	return signature, err
}

// ASCIIParser returns a hexadecimal representation of a signature
// written using ASCII encoding.
func ASCIIParser(signature string) string {
	return strings.ToUpper(hex.EncodeToString([]byte(signature)))
}

// PERLParser will take a very limited range of PERL syntax and convert
// it to something PRONOM compatible.
func PERLParser(signature string) (string, bool) {
	const perlSpace = "\\x20"
	const perlWildcardFour = ".{4}"
	if strings.Contains(signature, perlSpace) {
		signature = strings.Replace(signature, perlSpace, " ", 1)
	}
	if strings.Contains(signature, perlWildcardFour) {
		split := strings.Split(signature, perlWildcardFour)
		if len(split) == 2 {
			s1 := strings.ToUpper(hex.EncodeToString([]byte(split[0])))
			s2 := strings.ToUpper(hex.EncodeToString([]byte(split[1])))
			return fmt.Sprintf("%s{4}%s", s1, s2), true
		}
	}
	return "", false
}
