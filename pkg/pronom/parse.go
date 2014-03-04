package pronom

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/richardlehane/siegfried/pkg/core/bytematcher"
)

const (
	bofstring = "Absolute from BOF"
	eofstring = "Absolute from EOF"
	varstring = "Variable"
)

// an intermediary structure before creating a bytematcher.Frame
type token struct {
	min, max int
	pat      bytematcher.Pattern
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

// parse hexstrings
func parseHex(puid, hx string) ([]token, int, int, error) {
	tokens := make([]token, 0, 10)
	var choice bytematcher.Choice // common bucket for stuffing choices into
	var rangeStart string
	var min, max int
	l := sigLex(puid, hx)
	for i := l.nextItem(); i.typ != itemEOF; i = l.nextItem() {
		switch i.typ {
		case itemError:
			return nil, 0, 0, errors.New(i.String())
		// parse simple types
		case itemText:
			tokens = append(tokens, token{min, max, bytematcher.Sequence(decodeHex(i.val))})
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
			choice = make(bytematcher.Choice, 0, 2)
		case itemTextChoice:
			choice = append(choice, bytematcher.Sequence(decodeHex(i.val)))
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

func appendSig(s1, s2 bytematcher.Signature, pos string) bytematcher.Signature {
	if len(s1) == 0 {
		return s2
	}
	if pos == eofstring {
		return append(s1, s2...)
	}
	for i, f := range s1 {
		orientation := f.Orientation()
		if orientation == bytematcher.SUCC || orientation == bytematcher.EOF {
			s3 := make(bytematcher.Signature, len(s1)+len(s2))
			copy(s3, s1[:i])
			copy(s3[i:], s2)
			copy(s3[i+len(s2):], s1[i:])
			return s3
		}
	}
	return append(s1, s2...)
}

func (p *pronom) Parse() ([]bytematcher.Signature, error) {
	sigs := make([]bytematcher.Signature, 0, 700)
	for _, f := range p.droid.FileFormats {
		puid := f.Puid
		for _, s := range f.Signatures {
			sig := make(bytematcher.Signature, 0, 1)
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
				if max == 0 {
					max = min
				}
				// parse the hexstring
				toks, lmin, lmax, err := parseHex(puid, bs.Hex)
				if err != nil {
					return nil, err
				}
				// create a new signature for this set of tokens
				tokSig := make(bytematcher.Signature, len(toks))
				// check position and add patterns to signature
				switch bs.Position {
				case bofstring:
					if toks[0].min == 0 && toks[0].max == 0 {
						toks[0].min, toks[0].max = min, max
					}
					tokSig[0] = bytematcher.NewFrame(bytematcher.BOF, toks[0].pat, toks[0].min, toks[0].max)
					if len(toks) > 1 {
						for i, tok := range toks[1:] {
							tokSig[i+1] = bytematcher.NewFrame(bytematcher.PREV, tok.pat, tok.min, tok.max)
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
					tokSig[0] = bytematcher.NewFrame(bytematcher.BOF, toks[0].pat, toks[0].min, toks[0].max)
					if len(toks) > 1 {
						for i, tok := range toks[1:] {
							tokSig[i+1] = bytematcher.NewFrame(bytematcher.PREV, tok.pat, tok.min, tok.max)
						}
					}
				case eofstring:
					if len(toks) > 1 {
						for i, tok := range toks[:len(toks)-1] {
							tokSig[i] = bytematcher.NewFrame(bytematcher.SUCC, tok.pat, toks[i+1].min, toks[i+1].max)
						}
					}
					// handle edge case where there is a {x-y} at end of EOF seq e.g. x-fmt/263
					if lmin != 0 || lmax != 0 {
						min, max = lmin, lmax
					}
					tokSig[len(toks)-1] = bytematcher.NewFrame(bytematcher.EOF, toks[len(toks)-1].pat, min, max)
				default:
					return nil, errors.New("Pronom parse error: invalid ByteSequence position " + bs.Position)
				}
				// add the segment (tokens signature) to the complete signature
				sig = appendSig(sig, tokSig, bs.Position)
			}
			sigs = append(sigs, sig)
		}
	}
	return sigs, nil
}
