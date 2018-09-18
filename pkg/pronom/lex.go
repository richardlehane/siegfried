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
	itemCurlyLeft
	itemCurlyRight
	itemWildStart
	itemSlash
	itemWildEnd
	itemWildSingle //??
	itemWild       //*
	itemUnprocessedText
	itemEnterGroup
	itemExitGroup
	itemChoiceMarker
	itemNotMarker
	itemRangeMarker
	itemMaskMarker
	itemAnyMaskMarker
	itemHexText
	itemQuoteText
	itemQuote
	itemSpace
)

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
	amp          = '&'
	tilda        = '~'
	newline      = '\n'
	carriage     = '\r'
)

const digits = "0123456789"

const hexadecimal = digits + "abcdefABCDEF"

const hexnonquote = hexadecimal + " " + "\n" + "\r"

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

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// acceptText consumes a run of runes that are deemed to be plain sequences (hex or quoted values)
func (l *lexer) acceptText(group bool) error {
	valid := hexnonquote
	if group {
		valid = hexadecimal
	}
	for {
		l.acceptRun(valid)
		switch l.peek() {
		default:
			return nil
		case quot:
			r := l.next()
			for r = l.next(); r != eof && r != quot; r = l.next() {
			}
			if r != quot {
				return fmt.Errorf("expected closing quote, got %v", r)
			}
		}
	}
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

// lexer for PRONOM signature files - reports, container and droid
func lexPRONOM(name, input string) *lexer {
	return lex(name, input, insideText)
}

func insideText(l *lexer) stateFn {
	if err := l.acceptText(false); err != nil {
		return l.errorf(err.Error())
	}
	if l.pos > l.start {
		l.emit(itemUnprocessedText)
	}
	r := l.next()
	switch r {
	default:
		return l.errorf("encountered invalid character %q", r)
	case eof:
		l.emit(itemEOF)
		return nil
	case leftBracket:
		l.emit(itemEnterGroup)
		return insideLeftBracket
	case leftParens:
		l.emit(itemEnterGroup)
		return insideLeftParens
	case leftCurly:
		l.emit(itemCurlyLeft)
		return insideWild
	case wildSingle:
		return insideWildSingle
	case wild:
		l.emit(itemWild)
		return insideText
	}
}

func (l *lexer) insideGroup(boundary itemType) stateFn {
	depth := 1
	for {
		if err := l.acceptText(true); err != nil {
			return l.errorf(err.Error())
		}
		if l.pos > l.start {
			l.emit(itemUnprocessedText)
		}
		r := l.next()
		switch r {
		default:
			return l.errorf("encountered invalid character %q", r)
		case leftBracket:
			l.emit(itemEnterGroup)
			depth++
		case rightBracket:
			l.emit(itemExitGroup)
			depth--
			if depth == 0 {
				if boundary != rightBracket {
					return l.errorf("expected group to close with %q, got %q", boundary, r)
				}
				return insideText
			}
		case rightParens:
			if boundary != rightParens {
				return l.errorf("expected group to close with %q, got %q", boundary, r)
			}
			l.emit(itemExitGroup)
			return insideText
		case not:
			l.emit(itemNotMarker)
		case pipe, space, tab:
			l.emit(itemChoiceMarker)
		case colon, slash:
			l.emit(itemRangeMarker)
		case amp:
			l.emit(itemMaskMarker)
		case tilda:
			l.emit(itemAnyMaskMarker)
		}
	}
}

func insideLeftBracket(l *lexer) stateFn {
	return l.insideGroup(rightBracket)
}

func insideLeftParens(l *lexer) stateFn {
	return l.insideGroup(rightParens)
}

func insideWildSingle(l *lexer) stateFn {
	r := l.next()
	if r == wildSingle {
		l.emit(itemWildSingle)
		return insideText
	}
	return l.errorf("expecting a double '?', got %q", r)
}

func insideWild(l *lexer) stateFn {
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
		return insideText
	}
	return l.errorf("expecting a closing bracket, got %q", r)
}

// text lexer
func lexText(input string) *lexer {
	return lex("textProcessor", input, insideUnprocessedText)
}

func insideUnprocessedText(l *lexer) stateFn {
	for {
		l.acceptRun(hexadecimal)
		if l.pos > l.start {
			l.emit(itemHexText)
		}
		switch l.next() {
		default:
			l.backup()
			return l.errorf("unexpected character in text: %q", l.next())
		case eof:
			l.emit(itemEOF)
			return nil
		case quot:
			l.emit(itemQuote)
			return insideQuoteText
		case space, tab, newline, carriage:
			l.emit(itemSpace)
		}
	}
}

func insideQuoteText(l *lexer) stateFn {
	r := l.next()
	for ; r != eof && r != quot; r = l.next() {
	}
	if r == quot {
		l.backup()
		l.emit(itemQuoteText)
		l.next()
		l.emit(itemQuote)
		return insideUnprocessedText
	}
	return l.errorf("expected closing quote, reached end of string")
}
