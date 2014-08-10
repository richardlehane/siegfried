package patterns

// Shared test sequences (exported so they can be used by the other bytematcher packages)
var TestSequences = []Sequence{
	Sequence{'t', 'e', 's', 't'},
	Sequence{'t', 'e', 's', 't'},
	Sequence{'t', 'e', 's', 't', 'y'},
	Sequence{'T', 'E', 'S', 'T'},
	Sequence{'T', 'E', 'S', 'T', 'Y'},
	Sequence{'Y', 'N', 'E', 'S', 'S'},
}

// Shared test choices (exported so they can be used by the other bytematcher packages)
var TestChoices = []Choice{
	Choice{TestSequences[0], TestSequences[2]},
	Choice{TestSequences[2], TestSequences[0]},
	Choice{TestSequences[4], TestSequences[5]},
	Choice{TestSequences[3]},
}
