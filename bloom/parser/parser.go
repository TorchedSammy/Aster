package parser

import (
	"fmt"
	"io"
	"github.com/TorchedSammy/Aster/bloom/token"
	"github.com/TorchedSammy/Aster/bloom/ast"
)

func Parse(r io.Reader) (nodes []ast.Node, err error) {
	lx := token.NewLexer(r)
	ops := []ast.Node{}

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
		tok, pos, lit := lx.Next()
		if tok == token.EOF {
			break
		}

		switch tok {
			case token.VAR:
				n := ast.Decl{
					Pos: pos,
				}

				tok, pos, lit = lx.Next()
				expectToken(tok, token.IDENT)
				n.Name = lit

				tok, pos, lit = lx.Next()
				expectToken(tok, token.ASSIGN)

				tok, pos, lit = lx.Next()
				switch tok {
					case token.STRING:
						n.Val = ast.Value{
							Pos: pos,
							Val: lit,
							Kind: ast.StringKind,
						}
					case token.NUMBER:
						n.Val = ast.Value{
							Pos: pos,
							Val: lit,
							Kind: ast.NumberKind,
						}
				}

				ops = append(ops, n)
			case token.IDENT:
				// In our aster script language, a call does not have
				// parens. It's like shell script, so if we have
				// an identifier and then a literal after we can assume
				// its a call
				n := ast.Call{
					Pos: pos,
					Name: lit,
					Arguments: []ast.Value{},
				}

				tok, pos, lit = lx.Next()
				switch tok {
					case token.ILLEGAL:
						return ops, fmt.Errorf("oopsie")
					case token.STRING:
						val := ast.Value{
							Pos: pos,
							Val: lit,
							Kind: ast.StringKind,
						}
						n.Arguments = append(n.Arguments, val)
					case token.NUMBER:
						val := ast.Value{
							Pos: pos,
							Val: lit,
							Kind: ast.NumberKind,
						}
						n.Arguments = append(n.Arguments, val)
					case token.VAR_REF:
						// #variable is a reference to a variable
						// untagged identifiers are commands
						val := ast.Value{
							Pos: pos,
							Val: lit,
							Kind: ast.VariableKind,
						}
						n.Arguments = append(n.Arguments, val)
				}
				ops = append(ops, n)
				case token.SWITCH:
					// @switch is a command line option
					// @switch=val assigns a value while the latter toggles
					// a boolean
					n := ast.SwitchAssign{
						Pos: pos,
						IsToggle: true,
					}
					t2, _, lit := lx.Next()
					expectToken(t2, token.IDENT)

					n.Name = lit
					ops = append(ops, n)
			/*
			
			case token.LPAREN:
				// check what the last node was
				switch ops[len(ops) - 1].(type) {
					case ast.Call:
						
				}
			*/
		}
	}

	return ops, nil
}

func expectToken(t token.Token, expected token.Token) {
	if t != expected {
		panic("unexpected tok " + t.String())
	}
}
