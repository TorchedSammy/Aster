package script

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Token int
const (
	EOF = iota
	ILLEGAL
	COMMENT

	IDENT // print
	NUMBER // 420
	STRING // "a"

	// Operators and Delimiters
	ASSIGN // =
	LPAREN // (
	RPAREN // )
	VAR_REF // #

	// keywords
	VAR
)

var tokenIdentMap = map[Token]string{
	EOF: "EOF",
	ILLEGAL: "ILLEGAL",

	IDENT: "IDENT",
	COMMENT: "COMMENT",
	NUMBER: "NUMBER",
	STRING: "STRING",

	ASSIGN: "ASSIGN",
	LPAREN: "LPAREN",
	RPAREN: "RPAREN",
	VAR_REF: "VAR_REF",

	VAR: "VAR",
}

func (t Token) String() string {
	name := tokenIdentMap[t]
	if name == "" {
		return "UNKNOWN"
	}

	return name
}

type Position struct {
	Line int
	Column int
}

type Lexer struct {
	pos Position
	reader *bufio.Reader
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		pos: Position{Line: 1, Column: 0},
		reader: bufio.NewReader(reader),
	}
}

func (l *Lexer) Next() (Token, Position, string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return EOF, l.pos, ""
			}

			panic(err) // ?
		}

		l.pos.Column++

		switch r {
			case '\n':
				// do things with newLine
				l.pos.Line++
				l.pos.Column = 0
			case '"':
				start := l.pos
				lit, err := l.scanString()
				if err != nil {
					return ILLEGAL, start, ""
				}

				return STRING, start, lit
			case '=':
				return ASSIGN, l.pos, string(r)
			case '#':
				start := l.pos
				ident := l.scanIdent()

				return VAR_REF, start, ident
			default:
				if unicode.IsLetter(r) {
					start := l.pos
					l.Back() // to rescan part of the ident in the method below
					ident := l.scanIdent()

					if ident == "var" {
						return VAR, start, ident
					}

					return IDENT, start, ident
				} else if unicode.IsNumber(r) {
					start := l.pos
					l.Back()
					num := l.scanNumber()

					return NUMBER, start, num
				}
		}
	}
}

func (l *Lexer) Back() {
	l.reader.UnreadRune()
	l.pos.Column--
}

func (l *Lexer) readRune() (rune, bool, error) {
	r, _, err := l.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			return r, true, nil
		}

		return r, true, err
	}

	return r, false, nil
}

func (l *Lexer) scanIdent() string {
	sb := strings.Builder{}

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return sb.String()
			}
		}

		l.pos.Column++

		if unicode.IsLetter(r) {
			sb.WriteRune(r)
			continue
		}

		l.Back() // unread non-ident rune
		return sb.String()
	}
}

func (l *Lexer) scanString() (literal string, err error) {
	sb := strings.Builder{}
	escaped := false

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Unknown error")
			}
		}
	}()

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return sb.String(), nil
			}
		}
		l.pos.Column++

		switch r {
			case '\\':
				if !escaped {
					escaped = true
					continue
				}
				sb.WriteRune(r)
				escaped = false
			case '"':
				if !escaped {
					return sb.String(), nil // we're done
				}
				sb.WriteRune(r)
				escaped = false
			case 'x':
				if !escaped {
					sb.WriteRune(r)
					 continue
				}
				b := strings.Builder{}
				r1 := l.expectRune(isHex)
				r2 := l.expectRune(isHex)
				b.WriteRune(r1)
				b.WriteRune(r2)

				i, _ := strconv.ParseInt(b.String(), 16, 64)
				sb.WriteRune(rune(i))
				escaped = false
			default:
				sb.WriteRune(r)
		}
	}
}

func (l *Lexer) scanNumber() string {
	sb := strings.Builder{}

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return sb.String()
			}
		}
		l.pos.Column++

		if unicode.IsNumber(r) {
			sb.WriteRune(r)
			continue
		}

		return sb.String()
	}
}

func isHex(r rune) bool {
	return ('0' <= r && r <= '9') || ('a' <= r && r <= 'f')
}

func (l *Lexer) expectRune(cond func(rune) bool) rune {
	r, done, _ := l.readRune()
	if done {
		panic(fmt.Errorf("unexpected EOF"))
	}

	if !cond(r) {
		panic(fmt.Errorf("..."))
	}

	return r
}
