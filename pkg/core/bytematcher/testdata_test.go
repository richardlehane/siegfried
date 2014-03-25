package bytematcher

import (
	"sync"

	"github.com/richardlehane/siegfried/pkg/core/siegreader"

	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	. "github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
)

// Stubs used by multiple test files within the bytematcher package

// Pattern
var (
	seqStub     = Sequence{'t', 'e', 's', 't', 'y'}
	seqStub2    = Sequence{'t', 'e', 's', 't', 'y'}
	seqStub3    = Sequence{'T', 'E', 'S', 'T'}
	seqStub4    = Sequence{'T', 'E', 'S', 'T', 'Y'}
	seqStub5    = Sequence{'Y', 'N', 'E', 'S', 'S'}
	choiceStub  = Choice{Sequence{'t', 'e', 's', 't', 'y'}, Sequence{'t', 'e', 's', 't'}}
	choiceStub2 = Choice{Sequence{'t', 'e', 's', 't'}, Sequence{'t', 'e', 's', 't', 'y'}}
	choiceStub3 = Choice{seqStub3, seqStub4}
	choiceStub4 = Choice{seqStub3}
)

//Frame
var (
	fixedStub   = Fixed{BOF, 0, seqStub}
	fixedStub2  = Fixed{BOF, 0, seqStub3}
	fixedStub3  = Fixed{SUCC, 0, seqStub3}
	fixedStub4  = Fixed{PREV, 0, seqStub5}
	fixedStub5  = Fixed{SUCC, 1, seqStub}
	windowStub  = Window{BOF, 0, 5, seqStub}
	windowStub2 = Window{PREV, 10, 20, seqStub}
	windowStub3 = Window{EOF, 10, 20, choiceStub4}
	windowStub4 = Window{PREV, 0, 1, seqStub3}
	wildStub    = Wild{BOF, seqStub}
	wildStub2   = Wild{SUCC, choiceStub3}
	wildMinStub = WildMin{BOF, 5, seqStub}
)

//Signature
var (
	sigStub  Signature = Signature{fixedStub2, windowStub2, wildStub2, fixedStub3, windowStub3}
	sigStub2 Signature = Signature{fixedStub2, windowStub2, windowStub4, fixedStub4, wildStub2}
)

// Tests
var (
	testNodeStub *testNode = &testNode{
		Frame:   fixedStub3,
		Success: []int{},
		Tests:   []*testNode{testNodeStub2},
	}
	testNodeStub2 *testNode = &testNode{
		Frame:   fixedStub5,
		Success: []int{0},
		Tests:   []*testNode{},
	}
	testNodeStub3 *testNode = &testNode{
		Frame:   windowStub2,
		Success: []int{},
		Tests:   []*testNode{testNodeStub4},
	}
	testNodeStub4 *testNode = &testNode{
		Frame:   fixedStub4,
		Success: []int{0},
		Tests:   []*testNode{},
	}
)

//TestTree
var (
	testTreeStub *testTree = &testTree{
		Complete: []keyframeID{},
		Incomplete: []followUp{
			followUp{
				Kf: keyframeID{1, 0},
				L:  true,
				R:  true,
			},
		},
		MaxLeftDistance:  10,
		MaxRightDistance: 30,
		Left:             []*testNode{testNodeStub},
		Right:            []*testNode{testNodeStub3},
	}
)

// Bytematcher
var bmStub *ByteMatcher = &ByteMatcher{
	Sigs: [][]keyFrame{
		[]keyFrame{},
		[]keyFrame{
			keyFrame{Typ: BOF, Min: 0, Max: 12},
		},
	},
	TestSet: []*testTree{testTreeStub},

	BofSeqs: &seqSet{},
	EofSeqs: &seqSet{},
	VarSeqs: &seqSet{},
}

// Matcher
var mStub = []byte{'t', 'e', 's', 't', 'y', 'A', 'T', 'E', 'S', 'T', 'M', 'A', 'T', 'C', 'H', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 't', 'e', 's', 't', 'y', 'Y', 'N', 'E', 'S', 'S'}
var matcherStub *matcher = &matcher{
	b:                bmStub,
	buf:              siegreader.New(),
	r:                make(chan int),
	partialKeyframes: make(map[[2]int][][2]int),
	limit:            nil,
	limitm:           &sync.RWMutex{},
	limitc:           nil,
	incoming:         make(chan strike),
	quit:             make(chan struct{}),
}

// Keyframes

var (
	kfstub = []keyFrame{
		keyFrame{
			Typ: BOF,
			Min: 8,
			Max: 12,
		},
		keyFrame{
			Typ: PREV,
			Min: 5,
			Max: 5,
		},
		keyFrame{
			Typ: PREV,
			Min: 0,
			Max: -1,
		},
		keyFrame{
			Typ: SUCC,
			Min: 5,
			Max: 10,
		},
		keyFrame{
			Typ: EOF,
			Min: 0,
			Max: 0,
		},
	}
)

// Partial keyframes

var (
	pstub = [][][2]int{
		[][2]int{
			[2]int{10, 5},
			[2]int{7, 2},
		},
		[][2]int{
			[2]int{20, 5},
			[2]int{20, 10},
		},
		[][2]int{
			[2]int{24, 5},
			[2]int{40, 5},
		},
		[][2]int{
			[2]int{50, 5},
		},
		[][2]int{
			[2]int{60, 10},
			[2]int{62, 8},
		},
	}
)
