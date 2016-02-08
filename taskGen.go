package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/bradfitz/slice"

	log "github.com/Sirupsen/logrus"
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

type Iterator interface {
	push(item)
	configure()
	cycle()
	get() string
	length() int
	finished() bool
	setCyclePos(int)
}

type TextIterator struct {
	text string
}

func (i *TextIterator) push(it item)    { i.text = it.val }
func (i *TextIterator) get() string     { return i.text }
func (i *TextIterator) cycle()          {}
func (i *TextIterator) configure()      {}
func (i *TextIterator) length() int     { return 1 }
func (i *TextIterator) finished() bool  { return true }
func (i *TextIterator) setCyclePos(int) {}

type ChoiceIterator struct {
	choices      []string
	currentCycle int
	tmpCycle     int
	cyclepos     int
}

func (i *ChoiceIterator) push(it item) {
	if it.typ == itemText {
		i.choices = append(i.choices, it.val)
	}
}
func (i *ChoiceIterator) cycle() {
	i.tmpCycle = (i.tmpCycle + 1) % i.cyclepos
	if i.tmpCycle == 0 {
		i.currentCycle = (i.currentCycle + 1) % i.length()
	}
}
func (i *ChoiceIterator) get() string         { return i.choices[i.currentCycle] }
func (i *ChoiceIterator) configure()          {}
func (i *ChoiceIterator) length() int         { return len(i.choices) }
func (i *ChoiceIterator) finished() bool      { return i.currentCycle == 0 && i.tmpCycle == 0 }
func (i *ChoiceIterator) setCyclePos(pos int) { i.cyclepos = pos }

type RangeIterator struct {
	items        []item
	fn           func(*RangeIterator) float64
	cycleLength  int
	currentCycle int
	tmpCycle     int
	cyclepos     int
}

func (i *RangeIterator) push(it item) {
	if it.typ == itemNumber || it.typ == itemRangeSep {
		i.items = append(i.items, it)
	}
}

func (i *RangeIterator) cycle() {
	i.tmpCycle = (i.tmpCycle + 1) % i.cyclepos
	if i.tmpCycle == 0 {
		i.currentCycle++
		if i.length() > 0 {
			i.currentCycle = i.currentCycle % i.length()
		}
	}
}

func (i *RangeIterator) get() string {
	return strconv.FormatFloat(i.fn(i), 'f', -1, 32) // FIXME
}

func (i *RangeIterator) configure() {
	chkParseError := func(i item) float64 {
		f, e := strconv.ParseFloat(i.val, 64)
		if e != nil {
			panic(fmt.Errorf("couldn't parse number: %v", i.val))
		}
		return f
	}

	switch len(i.items) {
	case 2:
		a, b := i.items[0], i.items[1]
		chkParseError(a)
		i.cycleLength = -1
		if b.typ == itemRangeSep {
			start := chkParseError(a)
			i.fn = func(i *RangeIterator) float64 {
				return start + float64(i.currentCycle)
			}
		} else {
			panic(fmt.Errorf("..x not supported as range"))
		}
	case 3:
		a := chkParseError(i.items[0])
		b := chkParseError(i.items[2])
		i.cycleLength = int(b - a)
		if i.cycleLength < 0 {
			i.cycleLength = -i.cycleLength
		}
		i.fn = func(i *RangeIterator) float64 {
			if a < b {
				return a + float64(i.currentCycle)
			}
			return a - float64(i.currentCycle)
		}
	case 5:
		a := chkParseError(i.items[0])
		b := chkParseError(i.items[2])
		c := chkParseError(i.items[4])
		i.cycleLength = int((c - a) / b)
		if i.cycleLength < 0 {
			i.cycleLength = -i.cycleLength
		}
		i.fn = func(i *RangeIterator) float64 {
			if a < c {
				return a + float64(i.currentCycle)*b
			}
			return a - float64(i.currentCycle)*b
		}
	default:
		panic(fmt.Errorf("invalid amount of range items: %d", len(i.items)))
	}
}
func (i *RangeIterator) length() int         { return i.cycleLength }
func (i *RangeIterator) finished() bool      { return i.currentCycle == 0 && i.tmpCycle == 0 }
func (i *RangeIterator) setCyclePos(pos int) { i.cyclepos = pos }

func generateTasks(optionGen string) {
	//optionGen = "--compute -x \\[0..0.2..10] -y \\(on|off) -kek \\(1|2|3) \\[ 1.. ]"
	//optionGen = "\\(one|two|3) \\(eins|zwei|drei)"
	log.WithField("options", optionGen).Info("generating tasks")
	_, items := lex(optionGen)
	var iterators []Iterator
	var currIter Iterator
	emit := func() {
		currIter.configure()
		iterators = append(iterators, currIter)
		currIter = nil
	}
	for item := range items {
		switch item.typ {
		case itemRange:
			currIter = &RangeIterator{}
		case itemChoice:
			currIter = &ChoiceIterator{}
		case itemText:
			if currIter == nil {
				currIter = &TextIterator{text: item.val}
				emit()
			} else {
				currIter.push(item)
			}
		case itemIterEnd:
			emit()
		case itemEOF:
		default:
			currIter.push(item)
		}
	}

	sortable := make([]Iterator, len(iterators))
	for i := 0; i < len(iterators); i++ {
		iterators[i].configure()
		sortable[i] = iterators[i]
	}
	slice.Sort(sortable[:], func(i, j int) bool {
		return sortable[i].length() < sortable[j].length() || sortable[j].length() == -1
	})
	cyclepos := 1
	for i := 0; i < len(sortable); i++ {
		sortable[i].setCyclePos(cyclepos)
		//fmt.Printf("%+v\n", sortable[i])
		cyclepos *= sortable[i].length()
	}

	for {
		s := ""
		for i := 0; i < len(iterators); i++ {
			s += iterators[i].get()
			iterators[i].cycle()
		}
		task := new(Task)
		task.id = getNextTaskID()
		task.options = s
		log.WithField("task", task).Info("generated task")
		task.reportStatus(protos.TaskStatus_REGISTERED)
		taskChan <- task
		if iterators[len(iterators)-1].finished() {
			return
		}
	}
}
