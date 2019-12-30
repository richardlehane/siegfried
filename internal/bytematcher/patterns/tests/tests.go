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

import . "github.com/richardlehane/siegfried/internal/bytematcher/patterns"

// TestSequences are exported so they can be used by the other bytematcher packages.
var TestSequences = []Sequence{
	Sequence("test"),
	Sequence("test"),
	Sequence("testy"),
	Sequence("TEST"),
	Sequence("TESTY"),
	Sequence("YNESS"), //5
	{'a'},
	{'b'},
	{'c'},
	{'d'},
	{'e'},
	{'f'},
	{'g'},
	{'h'},
	{'i'},
	{'j'},
	Sequence("junk"), // 16
	Sequence("23"),
}

// TestNotSequences are exported so they can be used by the other bytematcher packages.
var TestNotSequences = []Not{
	{Sequence("test")},
	{Sequence("test")},
	{Sequence{255}},
	{Sequence{0}},
	{Sequence{10}},
}

// TestLists are exported so they can be used by the other bytematcher packages.
var TestLists = []List{
	{TestSequences[0], TestSequences[2]},
	{TestSequences[3], TestSequences[4]},
}

// Test Choices are exported so they can be used by the other bytematcher packages.
var TestChoices = []Choice{
	{TestSequences[0], TestSequences[2]},
	{TestSequences[2], TestSequences[0]},
	{TestSequences[4], TestSequences[5]},
	{TestSequences[3]},
	{
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
	{TestSequences[0], TestLists[0]},
	{TestSequences[3], TestSequences[4]},
}

var TestMasks = []Mask{Mask(0xAA)}

var TestAnyMasks = []AnyMask{AnyMask(0xAA)}
