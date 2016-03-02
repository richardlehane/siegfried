// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"strings"
)

// Name of the default identifier as well as settings for how a new identifer will be built
var identifier = struct {
	name        string // Name of the default identifier
	details     string // a short string describing the signature e.g. with what DROID and container file versions was it built?
	maxBOF      int    // maximum offset from beginning of file to scan
	maxEOF      int    // maximum offset from end of file to scan
	noEOF       bool   // trim end of file segments from signatures
	noContainer bool   // don't build with container signatures
	noPriority  bool   // ignore priority relations between signatures
	noText      bool   // don't build with text signatures
	noName      bool   // don't build with filename signatures
	noMIME      bool   // don't build with MIME signatures
	noXML       bool   // don't build with XML signatures
}{
	name: "pronom",
}

// GETTERS

// Name returns the name of the identifier.
func Name() string {
	return identifier.name
}

// Details returns a description of the identifier. This is auto-populated if not set directly.
func Details() string {
	// if the details string has been explicitly set, return it
	if len(identifier.details) > 0 {
		return identifier.details
	}
	// ... otherwise create a default string based on the identifier settings chosen
	str := DroidBase()
	if !identifier.noContainer {
		str += "; " + ContainerBase()
	}
	if identifier.maxBOF > 0 {
		str += fmt.Sprintf("; max BOF %d", identifier.maxBOF)
	}
	if identifier.maxEOF > 0 {
		str += fmt.Sprintf("; max EOF %d", identifier.maxEOF)
	}
	if identifier.noEOF {
		str += "; no EOF signature parts"
	}
	if identifier.noContainer {
		str += "; no container signatures"
	}
	if identifier.noPriority {
		str += "; no priorities"
	}
	if identifier.noText {
		str += "; no text matcher"
	}
	if identifier.noName {
		str += "; no filename matcher"
	}
	if identifier.noMIME {
		str += "; no MIME matcher"
	}
	if identifier.noXML {
		str += "; no XML matcher"
	}
	if pronom.noreports {
		str += "; built without reports"
	}
	if pronom.doubleup {
		str += "; byte signatures included for formats that also have container signatures"
	}
	if HasLimit() {
		str += "; limited to puids: " + strings.Join(pronom.limit, ", ")
	}
	if HasExclude() {
		str += "; excluding puids: " + strings.Join(pronom.exclude, ", ")
	}
	if len(pronom.extend) > 0 {
		str += "; extensions: " + strings.Join(pronom.extend, ", ")
	}
	if len(pronom.extendc) > 0 {
		str += "; container extensions: " + strings.Join(pronom.extendc, ", ")
	}
	return str
}

// MaxBOF returns any BOF buffer limit set.
func MaxBOF() int {
	return identifier.maxBOF
}

// MaxEOF returns any EOF buffer limit set.
func MaxEOF() int {
	return identifier.maxEOF
}

// NoEOF reports whether end of file segments of signatures should be trimmed.
func NoEOF() bool {
	return identifier.noEOF
}

// NoContainer reports whether container signatures should be omitted.
func NoContainer() bool {
	return identifier.noContainer
}

// NoPriority reports whether priorities between signatures should be omitted.
func NoPriority() bool {
	return identifier.noPriority
}

// NoText reports whether text signatures should be omitted.
func NoText() bool {
	return identifier.noText
}

// NoName reports whether filename signatures should be omitted.
func NoName() bool {
	return identifier.noName
}

// NoMIME reports whether MIME signatures should be omitted.
func NoMIME() bool {
	return identifier.noMIME
}

// NoXML reports whether XML signatures should be omitted.
func NoXML() bool {
	return identifier.noXML
}

// SETTERS

// SetName sets the name of the identifier.
func SetName(n string) func() private {
	return func() private {
		identifier.name = n
		return private{}
	}
}

// SetDetails sets the identifier's description. If not provided, this description is
// automatically generated based on options set.
func SetDetails(d string) func() private {
	return func() private {
		identifier.details = d
		return private{}
	}
}

// SetBOF limits the number of bytes to scan from the beginning of file.
func SetBOF(b int) func() private {
	return func() private {
		identifier.maxBOF = b
		return private{}
	}
}

// SetEOF limits the number of bytes to scan from the end of file.
func SetEOF(e int) func() private {
	return func() private {
		identifier.maxEOF = e
		return private{}
	}
}

// SetNoEOF will cause end of file segments to be trimmed from signatures.
func SetNoEOF() func() private {
	return func() private {
		identifier.noEOF = true
		return private{}
	}
}

// SetNoContainer will cause container signatures to be omitted.
func SetNoContainer() func() private {
	return func() private {
		identifier.noContainer = true
		return private{}
	}
}

// SetNoPriority will cause priority relations between signatures to be omitted.
func SetNoPriority() func() private {
	return func() private {
		identifier.noPriority = true
		return private{}
	}
}

// SetNoText will cause text signatures to be omitted.
func SetNoText() func() private {
	return func() private {
		identifier.noText = true
		return private{}
	}
}

// SetNoName will cause extension signatures to be omitted.
func SetNoName() func() private {
	return func() private {
		identifier.noName = true
		return private{}
	}
}

// SetNoMIME will cause MIME signatures to be omitted.
func SetNoMIME() func() private {
	return func() private {
		identifier.noMIME = true
		return private{}
	}
}

// SetNoXML will cause MIME signatures to be omitted.
func SetNoXML() func() private {
	return func() private {
		identifier.noXML = true
		return private{}
	}
}
