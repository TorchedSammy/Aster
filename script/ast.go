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

func Parse(r io.Reader) ([]Node, error) {
	lx := NewLexer(r)

	// TODO: better error handling
	// and definitely code cleanup
	ops := []Node{}
	for {
		token, pos, lit := lx.Next()
		if token == EOF {
			break
		}

		fmt.Printf("%s %s\n", token, lit)
		switch token {
			case VAR:
				n := Decl{
					Pos: pos,
				}

				token, pos, lit = lx.Next()
				fmt.Printf("%s %s\n", token, lit)
				if token != IDENT {
					return ops, fmt.Errorf("bad decl") 
				}
				n.Name = lit

				token, pos, lit = lx.Next()
				fmt.Printf("%s %s\n", token, lit)
				if token != ASSIGN {
					return ops, fmt.Errorf("bad decl")
				}

				token, pos, lit = lx.Next()
				fmt.Printf("%s %s\n", token, lit)
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
				fmt.Printf("%s %s\n", token, lit)
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
			/*
			
			case LPAREN:
				// check what the last node was
				switch ops[len(ops) - 1].(type) {
					case Call:
						
				}
			*/
		}
	}

	fmt.Printf("Nodes: %+q\n", ops)
	return ops, nil
}
