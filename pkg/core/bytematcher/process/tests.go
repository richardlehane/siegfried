package process

import (
	"github.com/richardlehane/match/wac"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
)

// Shared test keyFrames (exported so they can be used by the other bytematcher packages)
var TestKeyFrames = []keyFrame{
	keyFrame{
		Typ: frames.BOF,
		Seg: keyFramePos{
			PMin: 8,
			PMax: 12,
		},
	},
	keyFrame{
		Typ: frames.PREV,
		Seg: keyFramePos{
			PMin: 5,
			PMax: 5,
		},
	},
	keyFrame{
		Typ: frames.PREV,
		Seg: keyFramePos{
			PMin: 0,
			PMax: -1,
		},
	},
	keyFrame{
		Typ: frames.SUCC,
		Seg: keyFramePos{
			PMin: 5,
			PMax: 10,
		},
	},
	keyFrame{
		Typ: frames.EOF,
		Seg: keyFramePos{
			PMin: 0,
			PMax: 0,
		},
	},
}

// Shared test testNodes (exported so they can be used by the other bytematcher packages)
var TesttestNodes = []*testNode{
	&testNode{
		Frame:   frames.TestFrames[3],
		Success: []int{},
		Tests: []*testNode{
			&testNode{
				Frame:   frames.TestFrames[1],
				Success: []int{0},
				Tests:   []*testNode{},
			},
		},
	},
	&testNode{
		Frame:   frames.TestFrames[6],
		Success: []int{},
		Tests: []*testNode{
			&testNode{
				Frame:   frames.TestFrames[2],
				Success: []int{0},
				Tests:   []*testNode{},
			},
		},
	},
}

// Shared test testTree (exported so they can be used by the other bytematcher packages)
var TestTestTree = &testTree{
	Complete: []KeyFrameID{},
	Incomplete: []FollowUp{
		FollowUp{
			Kf: KeyFrameID{1, 0},
			L:  true,
			R:  true,
		},
	},
	MaxLeftDistance:  10,
	MaxRightDistance: 30,
	Left:             []*testNode{TesttestNodes[0]},
	Right:            []*testNode{TesttestNodes[1]},
}

var TestSeqSetBof = &seqSet{
	Set:           []wac.Seq{},
	TestTreeIndex: []int{},
}

var TestSeqSetEof = &seqSet{
	Set:           []wac.Seq{},
	TestTreeIndex: []int{},
}

var TestFrameSetBof = &frameSet{
	Set:           []frames.Frame{},
	TestTreeIndex: []int{},
}

var TestProcessObj = &Process{
	KeyFrames: [][]keyFrame{},
	Tests:     []*testTree{},
	BOFFrames: nil,
	EOFFrames: nil,
	BOFSeq:    nil,
	EOFSeq:    nil,
}

var Sample = []byte("testTESTMATCHAAAAAAAAAAAYNESStesty")
