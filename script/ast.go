package script

import (
	"fmt"
	"io"
)

type Node interface {
	Start() Position
	End() Position
}

type Decl struct{
	Name string
	Pos Position
	Val Value
}

func (d Decl) Start() Position {
	return d.Pos
}

func (d Decl) End() Position {
	return d.Val.End()
}

type BadDecl struct {
	Name string
}

type Expr struct{
	Pos Position
}

func (e Expr) Start() Position {
	return e.Pos
}

func (e Expr) End() Position {
	return e.Pos
}

type ParenExpr struct{
	LParen Position
	E Expr
	RParen Position
}

func (p ParenExpr) Start() Position {
	return p.LParen
}

func (p ParenExpr) End() Position {
	return p.RParen
}

type ValueKind int
const (
	EmptyKind ValueKind = iota
	StringKind
	NumberKind
	VariableKind
)

type Value struct{
	Pos Position
	Val string
	Kind ValueKind
}
var EmptyValue = Value{}

func (v Value) Start() Position {
	return v.Pos
}

func (v Value) End() Position {
	return v.Pos
}

func (v Value) String() string {
	return v.Val
}

type Call struct {
	Pos Position
	Name string // name of function
	Arguments []Value
}

func (c Call) Start() Position {
	return c.Pos
}

func (c Call) End() Position {
	return c.Pos
}

type SwitchAssign struct {
	Pos Position
	Name string
	Val Value
	IsToggle bool
}

func (s SwitchAssign) Start() Position {
	return s.Pos
}

func (s SwitchAssign) End() Position {
	return s.Pos
}

func Parse(r io.Reader) (nodes []Node, err error) {
	lx := NewLexer(r)
	ops := []Node{}

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Unknown error")
			}
		}
	}()

	// TODO: better error handling
	// and definitely code cleanup
	for {
		token, pos, lit := lx.Next()
		if token == EOF {
			break
		}

		switch token {
			case VAR:
				n := Decl{
					Pos: pos,
				}

				token, pos, lit = lx.Next()
				expectToken(token, IDENT)
				n.Name = lit

				token, pos, lit = lx.Next()
				expectToken(token, ASSIGN)

				token, pos, lit = lx.Next()
				switch token {
					case STRING:
						n.Val = Value{
							Pos: pos,
							Val: lit,
							Kind: StringKind,
						}
					case NUMBER:
						n.Val = Value{
							Pos: pos,
							Val: lit,
							Kind: NumberKind,
						}
				}

				ops = append(ops, n)
			case IDENT:
				// In our aster script language, a call does not have
				// parens. It's like shell script, so if we have
				// an identifier and then a literal after we can assume
				// its a call
				n := Call{
					Pos: pos,
					Name: lit,
					Arguments: []Value{},
				}

				token, pos, lit = lx.Next()
				switch token {
					case ILLEGAL:
						return ops, fmt.Errorf("oopsie")
					case STRING:
						val := Value{
							Pos: pos,
							Val: lit,
							Kind: StringKind,
						}
						n.Arguments = append(n.Arguments, val)
					case NUMBER:
						val := Value{
							Pos: pos,
							Val: lit,
							Kind: NumberKind,
						}
						n.Arguments = append(n.Arguments, val)
					case VAR_REF:
						// #variable is a reference to a variable
						// untagged identifiers are commands
						val := Value{
							Pos: pos,
							Val: lit,
							Kind: VariableKind,
						}
						n.Arguments = append(n.Arguments, val)
				}
				ops = append(ops, n)
				case SWITCH:
					// @switch is a command line option
					// @switch=val assigns a value while the latter toggles
					// a boolean
					n := SwitchAssign{
						Pos: pos,
						IsToggle: true,
					}
					t2, _, lit := lx.Next()
					expectToken(t2, IDENT)

					n.Name = lit
					ops = append(ops, n)
			/*
			
			case LPAREN:
				// check what the last node was
				switch ops[len(ops) - 1].(type) {
					case Call:
						
				}
			*/
		}
	}

	return ops, nil
}

func expectToken(t Token, expected Token) {
	if t != expected {
		panic("unexpected token " + t.String())
	}
}
