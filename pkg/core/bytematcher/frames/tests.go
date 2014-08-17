package frames

import "github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"

// Shared test frames (exported so they can be used by the other bytematcher packages)
var TestFrames = []Frame{
	Fixed{BOF, 0, patterns.TestSequences[0]},      //0 test
	Fixed{BOF, 0, patterns.TestSequences[1]},      // test
	Fixed{SUCC, 0, patterns.TestSequences[2]},     // testy
	Fixed{PREV, 0, patterns.TestSequences[3]},     // TEST
	Fixed{SUCC, 1, patterns.TestSequences[0]},     // test
	Window{BOF, 0, 5, patterns.TestSequences[0]},  //5 test
	Window{PREV, 10, 20, patterns.TestChoices[2]}, // TESTY | YNESS
	Window{EOF, 10, 20, patterns.TestChoices[0]},  // test | testy
	Window{PREV, 0, 1, patterns.TestSequences[3]}, // TEST
	Wild{BOF, patterns.TestSequences[0]},          // test
	Wild{SUCC, patterns.TestChoices[0]},           //10 test | testy
	WildMin{BOF, 5, patterns.TestSequences[0]},    // test
	WildMin{EOF, 5, patterns.TestSequences[0]},    // test
	Window{BOF, 0, 5, patterns.TestChoices[4]},    // a | b
	Wild{PREV, patterns.TestSequences[0]},         // test
	Wild{BOF, patterns.TestSequences[0]},          //15
	Wild{BOF, patterns.TestSequences[16]},
	Fixed{EOF, 0, patterns.TestSequences[17]},
}

// Shared test signatures (exported so they can be used by the other bytematcher packages)
var TestSignatures = []Signature{
	Signature{TestFrames[0], TestFrames[6], TestFrames[10], TestFrames[2], TestFrames[7]},                 // [BOF 0:test], [P 10-20:TESTY|YNESS], [S *:test|testy], [S 0:testy], [E 10-20:test|testy] 3 Segments
	Signature{TestFrames[1], TestFrames[6], TestFrames[8], TestFrames[2], TestFrames[10], TestFrames[17]}, // [BOF 0:test], [P 10-20:TESTY|YNESS], [P 0-1:TEST], [S 0:testy], [S *:test|testy], [E 0:23] 3 segments
	Signature{TestFrames[13], TestFrames[14]},                                                             // [BOF 0-5:a|b|c..j], [P *:test] 2 segments
	Signature{TestFrames[1], TestFrames[6], TestFrames[15]},                                               // [BOF 0:test], [P 10-20:TESTY|YNESS], [BOF *:test] 2 segments
	Signature{TestFrames[16]},                                                                             // [BOF *:junk]
}
