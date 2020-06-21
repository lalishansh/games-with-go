package apt

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type tokensType int

const (
	openParen tokensType = iota
	closeParen
	oprt
	constant
)
const eof rune = -1

type token struct {
	typ tokensType
	val string
}

type lexer struct {
	input  string
	start  int
	pos    int
	width  int
	tokens chan token
}

type stateFunc func(*lexer) stateFunc

func strToNode(s string) Node {
	switch s {
	case "+":
		return NewOpPlus()
	case "-":
		return NewOpMinus()
	case "*":
		return NewOpMult()
	case "/":
		return NewOpDiv()
	case "Sin":
		return NewOpSin()
	case "Cos":
		return NewOpCos()
	case "Atan":
		return NewOpAtan()
	case "Atan2":
		return NewOpAtan2()
	case "SimplexNoise":
		return NewOpNoise()
	case "Square":
		return NewOpSquare()
	case "Log2":
		return NewOpLog2()
	case "Ceil":
		return NewOpCeil()
	case "Floor":
		return NewOpFloor()
	case "Lerp":
		return NewOpLerp()
	case "Abs":
		return NewOpAbs()
	case "Clip":
		return NewOpClip()
	case "Wrap":
		return NewOpWrap()
	case "FBM":
		return NewOpFBM()
	case "Turbulence":
		return NewOpTurbulence()
	case "Negate":
		return NewOpNegate()
	case "X":
		return NewOpX()
	case "Y":
		return NewOpY()
	case "Picture":
		return NewOpPicture()
	default:
		panic("error in parser IN TOKEN STRING: " + s)
	}
}
func parse(tokens chan token, parent Node) Node {
	for {
		tokn, ok := <-tokens
		if !ok {
			panic("no more tokens")
		}
		switch tokn.typ {
		case oprt:
			n := strToNode(tokn.val)
			n.SetParent(parent)
			for i := range n.GetChildren() {
				n.GetChildren()[i] = parse(tokens, n)
			}
			return n
		case constant:
			n := NewOpConst()
			n.SetParent(parent)
			v, err := strconv.ParseFloat(tokn.val, 32)
			if err != nil {
				panic("error parsing float: Invalid Float present in File")
			}
			n.val = float32(v)
			return n
		case openParen, closeParen:
			continue
		}
	}
	return nil
}
func BeginLexing(s string) Node {
	l := &lexer{input: s, tokens: make(chan token, 100)}
	go l.run()
	return parse(l.tokens, nil)
}
func (l *lexer) run() {
	for state := determinTokens(l); state != nil; {
		state = state(l)
	}
	close(l.tokens)
}
func determinTokens(l *lexer) stateFunc {
	var r rune
	for {
		switch r = l.next(); {
		case isWhiteSpace(r):
			l.ignore()
		case r == '(':
			l.emit(openParen)
		case r == ')':
			l.emit(closeParen)
		case isStartOfNum(r):
			return lexNum
		case r == eof:
			return nil
		//case isOp(r):
		default:
			return lexOp
		}
	}
}
func lexOp(l *lexer) stateFunc {
	l.acceptRun("+-/*abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	l.emit(oprt)
	return determinTokens
}
func lexNum(l *lexer) stateFunc {
	l.accept("-.")
	digits := "1234567890"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.input[l.start:l.pos] == "-" {
		l.emit(oprt)
	} else {
		l.emit(constant)
	}
	return determinTokens
}
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}
func isWhiteSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t'
}
func isStartOfNum(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == '.'
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
func (l *lexer) backup() {
	l.pos -= l.width
}
func (l *lexer) ignore() {
	l.start = l.pos
}
func (l *lexer) emit(t tokensType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}
func (l *lexer) peek() (r rune) {
	r, _ = utf8.DecodeLastRuneInString(l.input[l.pos:])
	return r
}
