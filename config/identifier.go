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
	maxBOF      int
	maxEOF      int
	noEOF       bool
	noContainer bool
	noPriority  bool
}{
	name: "pronom",
}

// GETTERS

func Name() string {
	return identifier.name
}

func Details() string {
	// if the details string has been explicitly set, return it
	if len(identifier.details) > 0 {
		return identifier.details
	}
	// ... otherwise create a default string based on the identifier settings chosen
	str := fmt.Sprintf("signature v. %d; %s", siegfried.signatureVersion, DroidBase())
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
	if pronom.noreports {
		str += "; built from DROID signature, without using reports"
	}
	if HasInclude() {
		str += "; limited to puids: " + strings.Join(pronom.include, ", ")
	}
	if HasExclude() {
		str += "; excluding puids: " + strings.Join(pronom.exclude, ", ")
	}
	if len(pronom.extend) > 0 {
		str += "; extensions: " + strings.Join(pronom.extend, ", ")
	}
	return str
}

func MaxBOF() int {
	return identifier.maxBOF
}

func MaxEOF() int {
	return identifier.maxEOF
}

func NoEOF() bool {
	return identifier.noEOF
}

func NoContainer() bool {
	return identifier.noContainer
}

func NoPriority() bool {
	return identifier.noPriority
}

// SETTERS

func SetName(n string) func() private {
	return func() private {
		identifier.name = n
		return private{}
	}
}

func SetDetails(d string) func() private {
	return func() private {
		identifier.details = d
		return private{}
	}
}

func SetBOF(b int) func() private {
	return func() private {
		identifier.maxBOF = b
		return private{}
	}
}

func SetEOF(e int) func() private {
	return func() private {
		identifier.maxEOF = e
		return private{}
	}
}

func SetNoEOF() func() private {
	return func() private {
		identifier.noEOF = true
		return private{}
	}
}

func SetNoContainer() func() private {
	return func() private {
		identifier.noContainer = true
		return private{}
	}
}

func SetNoPriority() func() private {
	return func() private {
		identifier.noPriority = true
		return private{}
	}
}
