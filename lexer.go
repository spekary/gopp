package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
	"unicode"
)
const eof = -1
const (
	leftDelim    = "{"
	rightDelim   = "}"
	leftComment  = "/*"
	rightComment = "*/"
	lineComment = "//"
)

type Pos int


// reserved words and other tokens we care about
const (
	tokFunc = "func"
	tokClass = "class"
	tokExtends = "extends"
	tokPackage = "package"
)


//go:generate stringer -type=itemType
type itemType int

type item struct {
	typ itemType
	val string
}

const (
	itemError itemType = iota
	itemDot
	itemEOF
	itemClass
	itemExtends
	itemOpenBrace
	itemCloseBrace
	itemFunc
	itemText
	itemLeftDelim
	itemRightDelim
	itemFuncBody
	itemFuncParams
	itemMember
	itemComment
	itemLineComment
	itemPackage
)

func (i item) String() string {
	switch i.typ {
	case itemError:
		return i.val

	case itemEOF:
		return "EOF"

	}

	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}

	return fmt.Sprintf("%q", i.val)
}

type lexer struct {
	input string // string being scanned
	start int// start position of item
	pos int // current position
	lastPos int // last position of item read
	width int // width of last rune
	items chan item // channel of scanned items
}


type stateFn func(*lexer) stateFn


func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close (l.items)
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	item,ok := <-l.items

	if !ok {
		panic("Read past end of file")
	}

	l.lastPos = l.pos

	return item
}

func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func lexText(l *lexer) stateFn {
	for {
		l.acceptSpace();
		if strings.HasPrefix(l.input[l.pos:], tokClass + " ") {
			if (l.pos > l.start) {
				l.emit(itemText)	// emit text already read so far for straight output
			}
			return lexClass
		} else  if strings.HasPrefix(l.input[l.pos:], leftComment) {
			if (l.pos > l.start) {
				l.emit(itemText)	// emit text already read so far for straight output
			}
			return lexComment(l, lexText)
		} else  if strings.HasPrefix(l.input[l.pos:], tokPackage) {
			return lexIdentifier(l, itemPackage, lexText)
		}

		l.nextLine()

		if l.peek() == eof {
			break
		}

	}

	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil // stop
}

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer, nextState stateFn) stateFn {
	l.pos += int(len(leftComment))
	i := strings.Index(l.input[l.pos:], rightComment)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += int(i + len(rightComment))
	l.emit(itemComment)
	return nextState
}

func lexClass(l *lexer) stateFn {
	l.pos += len(tokClass)
	l.pos += 1 // consume space
	return lexIdentifier(l, itemClass, lexExtends)
}

func (l *lexer) next() rune {
	var c rune
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	c, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return c
}

func (l *lexer) nextLine() {
	for {
		r := l.next()
		if (r == '\n' || r == eof) {
			return
		}
	}
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) ignoreSpace() {
	for {
		r := l.next()
		switch  {
		case r == eof:
			return
		case isSpace(r):
			l.ignore()
		default:
			l.backup()
			return
		}
	}
}

func lexIdentifier(l *lexer, typ itemType, nextState stateFn) stateFn {
	var r rune

	l.ignore()
	l.ignoreSpace()

	for {
		switch r = l.next(); {
		case r == eof || r == '\n':
			return l.errorf("Missing identifier")
		case !isIdChar(r):
			if l.pos == l.start {
				return l.errorf("Invalid identifier")
			}
			l.backup()
			l.emit(typ)
			return nextState
		}
	}
}

// expecting "extends" keyword
func lexExtends(l *lexer) stateFn {
	l.ignoreSpace()
	if !strings.HasPrefix(l.input[l.pos:], tokExtends) {
		l.errorf("Missing 'extends' keyword")
	}
	l.pos += len(tokExtends)
	return lexExtendsClassName
}

func lexExtendsClassName(l *lexer) stateFn {
	return lexIdentifier(l, itemExtends, lexBodyOpen)
}

func lexBodyOpen(l *lexer) stateFn {
	//l.parenDepth = 0
	lexLeftDelim(l)
	return lexClassBody
}

func lexLeftDelim(l *lexer)  {
	l.ignoreSpace()
	l.pos += int(len(leftDelim))
	l.emit(itemLeftDelim)
	//l.parenDepth ++
}

func lexRightDelim(l *lexer)  {
	l.pos += int(len(rightDelim))
	l.emit(itemRightDelim)
	//l.parenDepth --
}



func lexClassBody(l *lexer) stateFn {
	l.ignoreSpace()
	if strings.HasPrefix(l.input[l.pos:], rightDelim) {
		return lexClassClose
	}

	if strings.HasPrefix(l.input[l.pos:], tokFunc + " ") {
		return lexFunc
	}

	if strings.HasPrefix(l.input[l.pos:], leftComment) {
		return lexComment(l, lexClassBody)
	}

	if strings.HasPrefix(l.input[l.pos:], lineComment) {
		l.nextLine()
		l.emit(itemLineComment)
		return lexClassBody
	}


	return lexMember
}

func lexClassClose(l *lexer) stateFn {
	lexRightDelim(l)
	return lexText
}


/**
Lex a function. We know the "func" keyword is at the beginning of the stream.
 */
func lexFunc(l *lexer) stateFn {
	l.pos += len (tokFunc)
	l.start = l.pos
	l.ignoreSpace()
	return lexIdentifier(l, itemFunc, lexFuncParams)
}

/**
Lex a function parameter list, including the return parameters. We need this because it will become part of the interface
definition and the struct definition.

 */
func lexFuncParams(l *lexer) stateFn {
	// TODO: Look for open parenthesis and close parenthesis first
	acceptUntil(l, leftDelim + "\n")
	if (l.peek() != '{') {
		return l.errorf("Missing opening bracket for function.")
	}
	l.emit(itemFuncParams)
	return lexFuncBody
}

/**
 */
func lexFuncBody(l *lexer) stateFn {
	// first find a left delim
	var r rune

	// TODO: Skip comments and quoted strings
	for r = l.next(); r != '{' && r != eof; r = l.next() {}
	var parenDepth = 1
	for parenDepth > 0 {
		r = l.next()
		if (r == '{') {
			parenDepth ++
		} else if (r == '}') {
			parenDepth --
		} else if (r == eof) {
			return l.errorf ("Unexpected EOF. Function body is still open.")
		}
	}
	l.emit(itemFuncBody)
	return lexClassBody
}

func lexMember(l *lexer) stateFn {
	l.nextLine()
	if (l.start + 1 < l.pos) {
		l.emit(itemMember);
	} else {
		l.ignore()
	}
	return lexClassBody
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isIdChar( reports whether r is an alphabetic, digit, underscore or period.
func isIdChar(r rune) bool {
	return r == '.' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func acceptUntil(l *lexer, terminators string) {
	for strings.IndexRune(terminators, l.next()) < 0 {
	}
	l.backup()
}

func  (l *lexer) acceptSpace() {
	for isSpace(l.next()) {
	}
	l.backup()
}


