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
)

// Name of the default identifier as well as settings for how a new identifer will be built
var Identifier = struct {
	Name        string // Name of the default identifier
	Details     string // a short string describing the signature e.g. with what DROID and container file versions was it built?
	MaxBOF      int
	MaxEOF      int
	NoEOF       bool
	NoContainer bool
	NoPriority  bool
}{
	Name: "pronom",
}

func Details() string {
	// if the details string has been explicitly set, return it
	if len(Identifier.Details) > 0 {
		return Identifier.Details
	}
	// ... otherwise create a default string based on the identifier settings chosen
	str := fmt.Sprintf("v%d; %s", Siegfried.SignatureVersion, Pronom.Droid)
	if !Identifier.NoContainer {
		str += "; " + Pronom.Container
	}
	if Identifier.MaxBOF > 0 {
		str += fmt.Sprintf("; max BOF %d", Identifier.MaxBOF)
	}
	if Identifier.MaxEOF > 0 {
		str += fmt.Sprintf("; max EOF %d", Identifier.MaxEOF)
	}
	if Identifier.NoEOF {
		str += "; no EOF signature parts"
	}
	if Identifier.NoContainer {
		str += "; no container signatures"
	}
	if Identifier.NoPriority {
		str += "; no priorities"
	}
	return str
}
