package strgen

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type itemType int
type item struct {
	typ itemType
	val string
}
type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan item
}

type stateFn func(*lexer) stateFn

const (
	itemError itemType = iota
	itemText
	itemNumber
	itemRange
	itemRangeSep
	itemChoice
	itemChoiceSep
	itemIterEnd
	itemDot
	itemEOF
)
const eof = -1

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	return fmt.Sprintf("%q", i.val)
}

func lex(input string) (*lexer, chan item) {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) skip(count int) {
	l.pos += count
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() (r rune) {
	r = l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for l.accept(valid) {
	}
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

func lexNumber(l *lexer) stateFn {
	l.accept("+-")
	digits := "0123456789"
	if !l.accept(digits) {
		return l.errorf("invalid digit")
	}
	l.acceptRun(digits)
	if l.accept(".") {
		if l.accept(".") {
			l.backup()
			l.backup()
			l.emit(itemNumber)
			l.skip(2)
			l.emit(itemRangeSep)
			return lexRange
		}
		l.acceptRun(digits)
	}
	l.emit(itemNumber)
	return lexRange
}

func lexRange(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "]") {
			l.next()
			l.emit(itemIterEnd)
			return lexText
		}
		switch r := l.next(); {
		case r == eof:
			return l.errorf("unclosed choice")
		case r == ' ':
			l.ignore()
		case r == '+' || r == '-' || '0' <= r && r <= '9':
			l.backup()
			return lexNumber
		case r == '.':
			if !l.accept(".") {
				return l.errorf("invalid range seperator")
			}
			l.emit(itemRangeSep)
		}
	}
}

func lexChoice(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], ")") {
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.next()
			l.emit(itemIterEnd)
			return lexText
		}
		switch r := l.next(); {
		case r == eof:
			return l.errorf("unclosed choice")
		case r == '|':
			l.backup()
			l.emit(itemText)
			l.next()
			l.emit(itemChoiceSep)
		}
	}
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "\\(") {
			l.emit(itemText)
			l.skip(2)
			l.emit(itemChoice)
			return lexChoice
		}
		if strings.HasPrefix(l.input[l.pos:], "\\[") {
			l.emit(itemText)
			l.skip(2)
			l.emit(itemRange)
			return lexRange
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}
