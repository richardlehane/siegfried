package pronom

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher/frames"
	"github.com/richardlehane/siegfried/pkg/core/bytematcher/patterns"
	"github.com/richardlehane/siegfried/pkg/pronom/mappings"
)

const (
	bofstring = "Absolute from BOF"
	eofstring = "Absolute from EOF"
	varstring = "Variable"
)

func (p *pronom) Parse() ([]frames.Signature, error) {
	sigs := make([]frames.Signature, 0, 700)
	for _, f := range p.droid.FileFormats {
		puid := f.Puid
		for _, s := range f.Signatures {
			sig, err := parseSig(puid, s)
			if err != nil {
				return nil, err
			}
			sigs = append(sigs, sig)
		}
	}
	return sigs, nil
}

// an intermediary structure before creating a bytematcher.Frame
type token struct {
	min, max int
	pat      patterns.Pattern
}

// helper funcs
func decodeHex(hx string) []byte {
	buf, _ := hex.DecodeString(hx) // ignore err, the hex string has been lexed
	return buf
}

func decodeNum(num string) (int, error) {
	if strings.TrimSpace(num) == "" {
		return 0, nil
	}
	return strconv.Atoi(num)
}

// parse hexstrings - puids are passed in for error reporting
func parseHex(puid, hx string) ([]token, int, int, error) {
	tokens := make([]token, 0, 10)
	var choice patterns.Choice // common bucket for stuffing choices into
	var rangeStart string
	var min, max int
	l := sigLex(puid, hx)
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemError:
			return nil, 0, 0, errors.New(i.String())
		// parse simple types
		case itemText:
			tokens = append(tokens, token{min, max, patterns.Sequence(decodeHex(i.val))})
		case itemNotText:
			tokens = append(tokens, token{min, max, NotSequence(decodeHex(i.val))})
		// parse range types
		case itemRangeStart, itemNotRangeStart, itemRangeStartChoice, itemNotRangeStartChoice:
			rangeStart = i.val
		case itemRangeEnd:
			tokens = append(tokens, token{min, max, Range{decodeHex(rangeStart), decodeHex(i.val)}})
		case itemNotRangeEnd:
			tokens = append(tokens, token{min, max, NotRange{decodeHex(rangeStart), decodeHex(i.val)}})
		// parse choice types
		case itemParensLeft:
			choice = make(patterns.Choice, 0, 2)
		case itemTextChoice:
			choice = append(choice, patterns.Sequence(decodeHex(i.val)))
		case itemNotTextChoice:
			choice = append(choice, NotSequence(decodeHex(i.val)))
		case itemRangeEndChoice:
			choice = append(choice, Range{decodeHex(rangeStart), decodeHex(i.val)})
		case itemNotRangeEndChoice:
			choice = append(choice, NotRange{decodeHex(rangeStart), decodeHex(i.val)})
		case itemParensRight:
			tokens = append(tokens, token{min, max, choice})
		// parse wild cards
		case itemWildSingle:
			min++
			max++
		case itemWildStart:
			min, _ = decodeNum(i.val)
		case itemCurlyRight: //detect {n} wildcards (i.e. not ranges) by checking if the max value has been set
			if max == 0 {
				max = min
			}
		case itemWildEnd:
			if i.val == "*" {
				max = -1
			} else {
				max, _ = decodeNum(i.val)
			}
		case itemWild:
			max = -1
		}
		// if we've appended a pattern, reset min and max
		switch i.typ {
		case itemText, itemNotText, itemRangeEnd, itemNotRangeEnd, itemParensRight:
			min, max = 0, 0
		}
	}
	return tokens, min, max, nil
}

// merge two segments into a signature. Provide s2's pos
func appendSig(s1, s2 frames.Signature, pos string) frames.Signature {
	if len(s1) == 0 {
		return s2
	}
	// if s2 is an EOF - just append it
	if pos == eofstring {
		return append(s1, s2...)
	}
	// if s1 already has an EOF segment, prepend that s2 segment before it, but after any preceding segments
	for i, f := range s1 {
		orientation := f.Orientation()
		if orientation == frames.SUCC || orientation == frames.EOF {
			s3 := make(frames.Signature, len(s1)+len(s2))
			copy(s3, s1[:i])
			copy(s3[i:], s2)
			copy(s3[i+len(s2):], s1[i:])
			return s3
		}
	}
	// default is just to append it
	return append(s1, s2...)
}

func parseSig(puid string, s mappings.Signature) (frames.Signature, error) {
	sig := make(frames.Signature, 0, 1)
	for _, bs := range s.ByteSequences {
		// check if <Offset> or <MaxOffset> elements are present
		min, err := decodeNum(bs.Offset)
		if err != nil {
			return nil, err
		}
		max, err := decodeNum(bs.MaxOffset)
		if err != nil {
			return nil, err
		}
		// lack of a max offset implies a fixed offset for BOF and EOF seqs (not VAR)
		if max == 0 {
			max = min
		}
		// parse the hexstring
		toks, lmin, lmax, err := parseHex(puid, bs.Hex)
		if err != nil {
			return nil, err
		}
		// create a new signature for this set of tokens
		tokSig := make(frames.Signature, len(toks))
		// check position and add patterns to signature
		switch bs.Position {
		case bofstring:
			if toks[0].min == 0 && toks[0].max == 0 {
				toks[0].min, toks[0].max = min, max
			}
			tokSig[0] = frames.NewFrame(frames.BOF, toks[0].pat, toks[0].min, toks[0].max)
			if len(toks) > 1 {
				for i, tok := range toks[1:] {
					tokSig[i+1] = frames.NewFrame(frames.PREV, tok.pat, tok.min, tok.max)
				}
			}
		case varstring:
			if max == 0 {
				max = -1
			}
			if toks[0].min == 0 && toks[0].max == 0 {
				toks[0].min, toks[0].max = min, max
			}
			if toks[0].min == toks[0].max {
				toks[0].max = -1
			}
			tokSig[0] = frames.NewFrame(frames.BOF, toks[0].pat, toks[0].min, toks[0].max)
			if len(toks) > 1 {
				for i, tok := range toks[1:] {
					tokSig[i+1] = frames.NewFrame(frames.PREV, tok.pat, tok.min, tok.max)
				}
			}
		case eofstring:
			if len(toks) > 1 {
				for i, tok := range toks[:len(toks)-1] {
					tokSig[i] = frames.NewFrame(frames.SUCC, tok.pat, toks[i+1].min, toks[i+1].max)
				}
			}
			// handle edge case where there is a {x-y} at end of EOF seq e.g. x-fmt/263
			if lmin != 0 || lmax != 0 {
				min, max = lmin, lmax
			}
			tokSig[len(toks)-1] = frames.NewFrame(frames.EOF, toks[len(toks)-1].pat, min, max)
		default:
			return nil, errors.New("Pronom parse error: invalid ByteSequence position " + bs.Position)
		}
		// add the segment (tokens signature) to the complete signature
		sig = appendSig(sig, tokSig, bs.Position)
	}
	return sig, nil
}

func parseContainerSeq(puid, seq string) ([]patterns.Pattern, error) {
	pats := make([]patterns.Pattern, 0, 10)
	var insideBracket bool
	var choiceMode bool
	var rangeMode bool
	var choice patterns.Choice // common bucket for stuffing choices into
	var firstBit []byte        // first text within brackets (could be a range or a choice)
	sequence := make(patterns.Sequence, 0)
	l := conLex(puid, seq)
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemError:
			return nil, errors.New(i.String())
		case itemText:
			if insideBracket {
				if choiceMode {
					choice = append(choice, patterns.Sequence(decodeHex(i.val)))
				} else if rangeMode {
					pats = append(pats, Range{firstBit, decodeHex(i.val)})
				} else {
					firstBit = decodeHex(i.val)
				}
			} else {
				sequence = append(sequence, decodeHex(i.val)...)
			}
		case itemQuoteText:
			if insideBracket {
				if choiceMode {
					choice = append(choice, patterns.Sequence(i.val))
				} else if rangeMode {
					pats = append(pats, Range{firstBit, []byte(i.val)})
				} else {
					firstBit = []byte(i.val)
				}
			} else {
				sequence = append(sequence, []byte(i.val)...)
			}
		case itemSpace:
			if insideBracket {
				if !choiceMode {
					choiceMode = true
					choice = patterns.Choice{patterns.Sequence(firstBit)}
				}
			}
		case itemSlash:
			if insideBracket {
				rangeMode = true
			} else {
				return nil, errors.New("Pronom parse error: unexpected slash in container (appears outside brackets)")
			}
		case itemColon:
			if insideBracket {
				rangeMode = true
			} else {
				return nil, errors.New("Pronom parse error: unexpected colon in container (appears outside brackets)")
			}
		case itemBracketLeft:
			if len(sequence) > 0 {
				pats = append(pats, sequence)
				sequence = make(patterns.Sequence, 0)
			}
			insideBracket = true
		case itemBracketRight:
			if choiceMode {
				pats = append(pats, choice)
			}
			insideBracket, choiceMode, rangeMode = false, false, false
		}
	}
	if len(sequence) > 0 {
		pats = append(pats, sequence)
	}
	return pats, nil
}

// Container signatures are simpler than regular Droid signatures
// No BOF/EOF/VAR - all are BOF.
// Min and Max Offsets usually provided. Lack of a Max Offset implies a Variable sequence.
// No wildcards within sequences: multiple subsequences with new offsets are used instead.
func parseContainerSig(puid string, s mappings.InternalSignature) (frames.Signature, error) {
	// some sigs only have paths, this is OK
	if s.ByteSequences == nil {
		return nil, nil
	}
	sig := make(frames.Signature, 0, 1)
	// Return an error for multiple byte sequences
	if len(s.ByteSequences) > 1 {
		return nil, errors.New("Pronom parse error: unexpected multiple byte sequences in container sig for puid " + puid)
	}
	bs := s.ByteSequences[0]
	// Return an error for non-BOF sequence
	if bs.Reference != "" && bs.Reference != "BOFoffset" {
		return nil, errors.New("Pronom parse error: unexpected reference in container sig for puid " + puid + "; bad reference is " + bs.Reference)
	}
	var prevPos int
	for i, sub := range bs.SubSequences {
		// Return an error if the positions don't increment.
		if sub.Position < prevPos {
			return nil, errors.New("Pronom parse error: container sub-sequences out of order for puid " + puid)
		}
		prevPos = sub.Position
		var typ frames.OffType
		if i == 0 {
			typ = frames.BOF
		} else {
			typ = frames.PREV
		}
		var min, max int
		min, _ = decodeNum(sub.SubSeqMinOffset)
		if sub.SubSeqMaxOffset == "" {
			max = -1
		} else {
			max, _ = decodeNum(sub.SubSeqMaxOffset)
		}
		pats, err := parseContainerSeq(puid, sub.Sequence)
		if err != nil {
			return nil, err
		}
		sig = append(sig, frames.NewFrame(typ, pats[0], min, max))
		if len(pats) > 1 {
			for _, v := range pats[1:] {
				sig = append(sig, frames.NewFrame(frames.PREV, v, 0, 0))
			}
		}
	}
	return sig, nil
}
