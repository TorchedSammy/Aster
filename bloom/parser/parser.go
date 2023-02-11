package parser

import (
	"fmt"
	"io"
	"github.com/TorchedSammy/Aster/bloom/token"
	"github.com/TorchedSammy/Aster/bloom/ast"
)

type Parser struct{
	lexer *token.Lexer
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(r io.Reader) (nodes []ast.Node, err error) {
	p.lexer = token.NewLexer(r)
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
		tok, pos, lit := p.lexer.Next()
		if tok == token.EOF {
			break
		}

		switch tok {
			case token.VAR:
				n := &ast.Decl{
					Pos: pos,
				}

				tok, pos, lit = p.lexer.Next()
				expectToken(tok, token.IDENT)
				n.Name = lit

				tok, pos, lit = p.lexer.Next()
				expectToken(tok, token.ASSIGN)

				tok, pos, lit = p.lexer.Next()
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
				if lit == "filter" {
					// defining filter
					n := p.Filter()
					ops = append(ops, n)
					continue
				}

				n := p.Command(lit, pos)
				ops = append(ops, n)
				case token.SWITCH:
					// @switch is a command line option
					// @switch=val assigns a value while the latter toggles
					// a boolean
					n := ast.SwitchAssign{
						Pos: pos,
						IsToggle: true,
					}
					t2, _, lit := p.lexer.Next()
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

// Command parses a command/function call.
func (p *Parser) Command(cmd string, cmdPos token.Position) *ast.Call {
	// In our aster script language, a call does not have
	// parens. It's like shell script, so if we have
	// an identifier and then a literal after we can assume
	// its a call
	var args []ast.Value

	for {
		t, pos, lit := p.lexer.Next()

		if pos.Line != cmdPos.Line || t == token.EOF {
			// TODO: unread token
			return &ast.Call{
				Name: cmd,
				Arguments: args,
			}
		}

		var val ast.Value
		switch t {
			case token.LBRACKET:
				// TODO: Parse bracket statements
				continue
			case token.RBRACKET:
				// TODO: same thing as above
				continue
			case token.IDENT:
				// TODO: identifiers in command calls have to be in paren
				// statements to avoid ambiguity: cause a parse error here
				continue
			case token.ILLEGAL:
				// TODO
			case token.STRING:
				val = ast.Value{
					Pos: pos,
					Val: lit,
					Kind: ast.StringKind,
				}
			case token.NUMBER:
				val = ast.Value{
					Pos: pos,
					Val: lit,
					Kind: ast.NumberKind,
				}
			case token.VAR_REF:
				// #variable is a reference to a variable
				// untagged identifiers are commands
				val = ast.Value{
					Pos: pos,
					Val: lit,
					Kind: ast.VariableKind,
				}
		}
		args = append(args, val)
	}
}

// Block parses a block.
func (p *Parser) Block() *ast.Block {
	lbr, lbrPos, _ := p.lexer.Next()
	expectToken(lbr, token.LBRACKET)

	var list []ast.Node
	for {
		t, pos, lit := p.next()
		switch t {
			case token.IDENT:
				list = append(list, p.Command(lit, pos))
			case token.RBRACKET:
				return &ast.Block{
					LBracket: lbrPos,
					List: list,
					RBracket: pos,
				}
			case token.EOF:
				panic("unexpected EOF")
		}
	}
}

// Filter parses a filter declaration.
// This is in the form of:
// filter name { ... }
func (p *Parser) Filter() *ast.FilterDeclaration {
	t1, pos, lit := p.next()
	expectToken(t1, token.IDENT) // name of filter

	n := &ast.FilterDeclaration{
		Pos: pos,
		Name: lit,
	}

	n.Body = p.Block()
	n.Body.InFilter = true

	return n
}

// next gets the next token and panics if it is EOF
func (p *Parser) next() (token.Token, token.Position, string) {
	t, pos, lit := p.lexer.Next()
	if t == token.EOF {
		panic("unexpected EOF")
	}

	return t, pos, lit
}

func expectToken(t token.Token, expected token.Token) {
	if t != expected {
		panic("unexpected tok " + t.String())
	}
}
