package patterns

// Shared test sequences (exported so they can be used by the other bytematcher packages)
var TestSequences = []Sequence{
	Sequence{'t', 'e', 's', 't'},
	Sequence{'t', 'e', 's', 't'},
	Sequence{'t', 'e', 's', 't', 'y'},
	Sequence{'T', 'E', 'S', 'T'},
	Sequence{'T', 'E', 'S', 'T', 'Y'},
	Sequence{'Y', 'N', 'E', 'S', 'S'}, //5
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
	Sequence{'j', 'u', 'n', 'k'}, // 16
	Sequence{'2', '3'},
}

// Shared test choices (exported so they can be used by the other bytematcher packages)
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
}
