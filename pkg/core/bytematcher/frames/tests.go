package frames

import "github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"

// Shared test frames (exported so they can be used by the other bytematcher packages)
var TestFrames = []Frame{
	Fixed{BOF, 0, patterns.TestSequences[0]},
	Fixed{BOF, 0, patterns.TestSequences[1]},
	Fixed{SUCC, 0, patterns.TestSequences[2]},
	Fixed{PREV, 0, patterns.TestSequences[3]},
	Fixed{SUCC, 1, patterns.TestSequences[0]},
	Window{BOF, 0, 5, patterns.TestSequences[0]},
	Window{PREV, 10, 20, patterns.TestChoices[2]},
	Window{EOF, 10, 20, patterns.TestChoices[0]},
	Window{PREV, 0, 1, patterns.TestSequences[3]},
	Wild{BOF, patterns.TestSequences[0]},
	Wild{SUCC, patterns.TestChoices[0]},
	WildMin{BOF, 5, patterns.TestSequences[0]},
}

// Shared test signatures (exported so they can be used by the other bytematcher packages)
var TestSignatures = []Signature{
	Signature{TestFrames[0], TestFrames[6], TestFrames[10], TestFrames[2], TestFrames[7]},
	Signature{TestFrames[1], TestFrames[6], TestFrames[8], TestFrames[3], TestFrames[10]},
}
