package script

import (
	"bufio"
	"io"
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
				return STRING, start, l.scanString()
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

func (l *Lexer) scanString() string {
	sb := strings.Builder{}
	escaped := false

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return sb.String()
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
					return sb.String() // we're done
				}
				sb.WriteRune(r)
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
