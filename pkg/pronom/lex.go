// The bulk of this lexer code is taken from http://golang.org/src/pkg/text/template/parse/lex.go
// Described in a talk by Rob Pike: http://cuddle.googlecode.com/hg/talk/lex.html#title-slide and http://www.youtube.com/watch?v=HxaD_trXwRE
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file. (Available here: http://golang.org/LICENSE)
//
// For the remainder of the file:
// Copyright 2014 Richard Lehane. All rights reserved.
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

// todo - a single lexing routine that will work for a single parse function for reports, droid and container

package pronom

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type item struct {
	typ itemType
	pos int
	val string
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	}
	return fmt.Sprintf("%q", i.val)
}

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemText //Sequence
	itemBracketLeft
	itemBracketRight
	itemNot
	itemNotText       //NotSequence
	itemNotRangeStart //NotRange
	itemNotRangeEnd   //NotRange
	itemColon
	itemRangeStart //Range
	itemRangeEnd   //Range
	itemParensLeft
	itemParensRight
	itemTextChoice          //Sequence
	itemRangeStartChoice    //Range
	itemRangeEndChoice      //Range
	itemNotTextChoice       //NotSequence
	itemNotRangeStartChoice //NotRange
	itemNotRangeEndChoice   //NotRange
	itemPipe
	itemCurlyLeft
	itemCurlyRight
	itemWildStart
	itemSlash
	itemWildEnd
	itemWildSingle //??
	itemWild       //*
	itemSpace
	itemQuote
	itemQuoteText
	itemCharText
)

const digits = "0123456789"

const hexadecimal = digits + "abcdefABCDEF"

const digitswild = digits + "*"

const eof = -1

type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name    string
	input   string
	state   stateFn
	pos     int
	start   int
	width   int
	lastPos int
	items   chan item
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == space || r == tab
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf("Lex error in "+l.name+": "+format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lexer for PRONOM signature files
func sigLex(name, input string) *lexer {
	return lex(name, input, sigText)
}

// lexer for TNA container files
func conLex(name, input string) *lexer {
	return lex(name, input, conText)
}

// lexer for DROID signature files
func droidLex(name, input string) *lexer {
	return lex(name, input, droidText)
}

// lex creates a new scanner for the input string.
func lex(name, input string, start stateFn) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run(start)
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run(start stateFn) {
	for l.state = start; l.state != nil; {
		l.state = l.state(l)
	}
}

const (
	leftBracket  = '['
	rightBracket = ']'
	leftParens   = '('
	rightParens  = ')'
	leftCurly    = '{'
	rightCurly   = '}'
	wildSingle   = '?'
	wild         = '*'
	not          = '!'
	colon        = ':'
	slash        = '-'
	pipe         = '|'
	quot         = '\''
	space        = ' '
	tab          = '\t'
)

// PRONOM signature lex states

func sigText(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.pos > l.start {
		l.emit(itemText)
	}
	r := l.next()
	switch r {
	case eof:
		l.emit(itemEOF)
		return nil
	case leftBracket:
		l.emit(itemBracketLeft)
		return sigLeftBracket
	case leftParens:
		l.emit(itemParensLeft)
		return sigInsideChoice
	case leftCurly:
		l.emit(itemCurlyLeft)
		return sigInsideWild
	case wildSingle:
		return sigWildSingle
	case wild:
		l.emit(itemWild)
		return sigText
	}
	return l.errorf("encountered invalid character %q", r)
}

func sigWildSingle(l *lexer) stateFn {
	r := l.next()
	if r == wildSingle {
		l.emit(itemWildSingle)
		return sigText
	}
	return l.errorf("expecting a double '?', got %q", r)
}

func sigLeftBracket(l *lexer) stateFn {
	if l.peek() == not {
		l.next()
		l.emit(itemNot)
		return sigNot
	}
	return sigInsideRange
}

func sigNot(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.peek() == colon {
		if l.pos > l.start {
			l.emit(itemNotRangeStart)
		}
		l.next()
		l.emit(itemColon)
		l.acceptRun(hexadecimal)
		if l.pos > l.start {
			l.emit(itemNotRangeEnd)
		}
	} else {
		if l.pos > l.start {
			l.emit(itemNotText)
		}
	}
	r := l.next()
	if r == rightBracket {
		l.emit(itemBracketRight)
		return sigText
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

func sigInsideRange(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.peek() == colon {
		if l.pos > l.start {
			l.emit(itemRangeStart)
		} else {
			return l.errorf("missing start value for range")
		}
		l.next()
		l.emit(itemColon)
	} else {
		return l.errorf("expecting a colon, got %q", l.peek())
	}
	l.acceptRun(hexadecimal)
	if l.pos > l.start {
		l.emit(itemRangeEnd)
	} else {
		return l.errorf("missing end value for range")
	}
	r := l.next()
	if r == rightBracket {
		l.emit(itemBracketRight)
		return sigText
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

func sigInsideChoice(l *lexer) stateFn {
	for {
		l.acceptRun(hexadecimal)
		if l.pos > l.start {
			l.emit(itemTextChoice)
		}
		r := l.next()
		switch r {
		case leftBracket:
			l.emit(itemBracketLeft)
			return sigLeftBracketChoice
		case not:
			l.emit(itemNot)
			return sigNotChoice
		case rightParens:
			l.emit(itemParensRight)
			return sigText
		case pipe:
			l.emit(itemPipe)
		default:
			return l.errorf("expecting a closing parens, got %q", r)
		}
	}
}

func sigLeftBracketChoice(l *lexer) stateFn {
	if l.peek() == not {
		l.next()
		l.emit(itemNot)
		return sigNotChoice
	}
	return sigInsideRangeChoice
}

func sigNotChoice(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.peek() == colon {
		if l.pos > l.start {
			l.emit(itemNotRangeStartChoice)
		}
		l.next()
		l.emit(itemColon)
		l.acceptRun(hexadecimal)
		if l.pos > l.start {
			l.emit(itemNotRangeEndChoice)
		}
	} else {
		if l.pos > l.start {
			l.emit(itemNotTextChoice)
			return sigInsideChoice
		}
	}
	r := l.next()
	if r == rightBracket {
		l.emit(itemBracketRight)
		return sigInsideChoice
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

func sigInsideRangeChoice(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.peek() == colon {
		if l.pos > l.start {
			l.emit(itemRangeStartChoice)
		}
		l.next()
		l.emit(itemColon)
	} else {
		l.errorf("expecting a colon, got %q", l.peek())
	}
	l.acceptRun(hexadecimal)
	if l.pos > l.start {
		l.emit(itemRangeEndChoice)
	}
	r := l.next()
	if r == rightBracket {
		l.emit(itemBracketRight)
		return sigInsideChoice
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

func sigInsideWild(l *lexer) stateFn {
	l.acceptRun(digits) // don't accept a '*' as start of range
	if l.pos > l.start {
		l.emit(itemWildStart)
	}
	r := l.next()
	if r == slash {
		l.emit(itemSlash)
		l.acceptRun(digitswild)
		if l.pos > l.start {
			l.emit(itemWildEnd)
		}
		r = l.next()
	}
	if r == rightCurly {
		l.emit(itemCurlyRight)
		return sigText
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

// Container signature lex states

func conText(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.pos > l.start {
		l.emit(itemText)
	}
	r := l.next()
	switch r {
	case eof:
		l.emit(itemEOF)
		return nil
	case leftBracket:
		l.emit(itemBracketLeft) // two types: [22 27] set & ['a'-'z'] range & : range
		return conText
	case rightBracket:
		l.emit(itemBracketRight)
		return conText
	case colon:
		l.emit(itemColon)
		return conText
	case slash:
		l.emit(itemSlash)
		return conText
	case quot:
		l.emit(itemQuote)
		return conInsideQuote
	case tab, space:
		l.emit(itemSpace)
		return conText
	}
	return l.errorf("encountered invalid character %q", r)
}

func conInsideQuote(l *lexer) stateFn {
	r := l.next()
	for ; r != eof && r != quot; r = l.next() {
	}
	if r == quot {
		l.backup()
		l.emit(itemQuoteText)
		l.next()
		l.emit(itemQuote)
		return conText
	}
	return l.errorf("expected closing quote, reached end of string")
}

// DROID signature lexer

func droidText(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.pos > l.start {
		l.emit(itemText)
	}
	r := l.next()
	switch r {
	case eof:
		l.emit(itemEOF)
		return nil
	case leftBracket:
		l.emit(itemBracketLeft)
		return droidLeftBracket
	}
	return l.errorf("encountered invalid character %q", r)
}

func droidLeftBracket(l *lexer) stateFn {
	if l.peek() == not {
		l.next()
		l.emit(itemNot)
		return droidNot
	}
	return droidInsideRange
}

func droidNot(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.peek() == colon {
		if l.pos > l.start {
			l.emit(itemNotRangeStart)
		}
		l.next()
		l.emit(itemColon)
		l.acceptRun(hexadecimal)
		if l.pos > l.start {
			l.emit(itemNotRangeEnd)
		}
	} else {
		if l.pos > l.start {
			l.emit(itemNotText)
		}
	}
	r := l.next()
	if r == rightBracket {
		l.emit(itemBracketRight)
		return droidText
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

func droidInsideRange(l *lexer) stateFn {
	l.acceptRun(hexadecimal)
	if l.peek() == colon {
		if l.pos > l.start {
			l.emit(itemRangeStart)
		} else {
			return l.errorf("missing start value for range")
		}
		l.next()
		l.emit(itemColon)
	} else {
		return l.errorf("expecting a colon, got %q", l.peek())
	}
	l.acceptRun(hexadecimal)
	if l.pos > l.start {
		l.emit(itemRangeEnd)
	} else {
		return l.errorf("missing end value for range")
	}
	r := l.next()
	if r == rightBracket {
		l.emit(itemBracketRight)
		return droidText
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}
