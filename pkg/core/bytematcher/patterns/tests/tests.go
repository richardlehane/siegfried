// Copyright 2015 Richard Lehane. All rights reserved.
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

// Package tests exports shared patterns for use by the other bytematcher packages
package tests

import . "github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"

// TestSequences are exported so they can be used by the other bytematcher packages.
var TestSequences = []Sequence{
	Sequence("test"),
	Sequence("test"),
	Sequence("testy"),
	Sequence("TEST"),
	Sequence("TESTY"),
	Sequence("YNESS"), //5
	Sequence{'a'},
	Sequence{'b'},
	Sequence{'c'},
	Sequence{'d'},
	Sequence{'e'},
	Sequence{'f'},
	Sequence{'g'},
	Sequence{'h'},
	Sequence{'i'},
	Sequence{'j'},
	Sequence("junk"), // 16
	Sequence("23"),
}

// TestNotSequences are exported so they can be used by the other bytematcher packages.
var TestNotSequences = []Not{
	Not{Sequence("test")},
	Not{Sequence("test")},
	Not{Sequence{255}},
	Not{Sequence{0}},
	Not{Sequence{10}},
}

// TestLists are exported so they can be used by the other bytematcher packages.
var TestLists = []List{
	List{TestSequences[0], TestSequences[2]},
	List{TestSequences[3], TestSequences[4]},
}

// Test Choices are exported so they can be used by the other bytematcher packages.
var TestChoices = []Choice{
	Choice{TestSequences[0], TestSequences[2]},
	Choice{TestSequences[2], TestSequences[0]},
	Choice{TestSequences[4], TestSequences[5]},
	Choice{TestSequences[3]},
	Choice{
		TestSequences[6],
		TestSequences[7],
		TestSequences[8],
		TestSequences[9],
		TestSequences[10],
		TestSequences[11],
		TestSequences[12],
		TestSequences[13],
		TestSequences[14],
		TestSequences[15],
	},
	Choice{TestSequences[0], TestLists[0]},
}
