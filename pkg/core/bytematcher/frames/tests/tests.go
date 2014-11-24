package tests

import (
	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns/tests"
)

// Shared test frames (exported so they can be used by the other bytematcher packages)
var TestFrames = []Frame{
	Fixed{BOF, 0, TestSequences[0]},      //0 test
	Fixed{BOF, 0, TestSequences[1]},      // test
	Fixed{SUCC, 0, TestSequences[2]},     // testy
	Fixed{PREV, 0, TestSequences[3]},     // TEST
	Fixed{SUCC, 1, TestSequences[0]},     // test
	Window{BOF, 0, 5, TestSequences[0]},  //5 test
	Window{PREV, 10, 20, TestChoices[2]}, // TESTY | YNESS
	Window{EOF, 10, 20, TestChoices[0]},  // test | testy
	Window{PREV, 0, 1, TestSequences[3]}, // TEST
	Wild{BOF, TestSequences[0]},          // test
	Wild{SUCC, TestChoices[0]},           //10 test | testy
	WildMin{BOF, 5, TestSequences[0]},    // test
	WildMin{EOF, 5, TestSequences[0]},    // test
	Window{BOF, 0, 5, TestChoices[4]},    // a | b
	Wild{PREV, TestSequences[0]},         // test
	Wild{BOF, TestSequences[0]},          //15
	Wild{BOF, TestSequences[16]},
	Fixed{EOF, 0, TestSequences[17]},
	Fixed{BOF, 0, TestLists[0]},
}

// Shared test signatures (exported so they can be used by the other bytematcher packages)
var TestSignatures = []Signature{
	Signature{TestFrames[0], TestFrames[6], TestFrames[10], TestFrames[2], TestFrames[7]},                 // [BOF 0:test], [P 10-20:TESTY|YNESS], [S *:test|testy], [S 0:testy], [E 10-20:test|testy] 3 Segments
	Signature{TestFrames[1], TestFrames[6], TestFrames[8], TestFrames[2], TestFrames[10], TestFrames[17]}, // [BOF 0:test], [P 10-20:TESTY|YNESS], [P 0-1:TEST], [S 0:testy], [S *:test|testy], [E 0:23] 3 segments
	Signature{TestFrames[13], TestFrames[14]},                                                             // [BOF 0-5:a|b|c..j], [P *:test] 2 segments
	Signature{TestFrames[1], TestFrames[6], TestFrames[15]},                                               // [BOF 0:test], [P 10-20:TESTY|YNESS], [BOF *:test] 2 segments
	Signature{TestFrames[16]},                                                                             // [BOF *:junk]
	Signature{TestFrames[18]},
}
