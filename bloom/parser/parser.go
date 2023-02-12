package parser

import (
	"fmt"
	"io"
	"runtime/debug"

	"github.com/TorchedSammy/Aster/bloom/token"
	"github.com/TorchedSammy/Aster/bloom/ast"
)

type Parser struct{
	lexer *token.Lexer
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(r io.Reader) (block *ast.Block, err error) {
	p.lexer = token.NewLexer(r)
	ops := []ast.Node{}

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Unknown error")
			}
			err = fmt.Errorf("%w\n%s", err, string(debug.Stack()))
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
				n := p.Decl(pos)
				ops = append(ops, n)
			case token.IDENT:
				n := p.Ident(lit, pos)

				ops = append(ops, n)
				case token.SWITCH:
					// @switch is a command line option
					// @switch=val assigns a value while the latter toggles
					// a boolean
					n := ast.SwitchAssign{
						Pos: pos,
						IsToggle: true,
					}
					_, _, lit := p.expectToken(token.IDENT)

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

	return &ast.Block{
		List: ops,
	}, nil
}

func (p *Parser) Ident(name string, pos token.Position) ast.Node {
	switch name {
		case "var":
			return p.Decl(pos)
		case "filter":
			return p.Filter()
		case "command":
			return p.CommandDecl()
		default:
			return p.Command(name, pos)
	}
}

func (p *Parser) Decl(pos token.Position) *ast.Decl {
	n := &ast.Decl{
		Pos: pos,
	}

	_, _, lit := p.expectToken(token.IDENT)
	n.Name = lit

	p.expectToken(token.ASSIGN)

	tok, pos, lit := p.lexer.Next()
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

	return n
}

// Command parses a command/function call.
func (p *Parser) Command(cmd string, cmdPos token.Position) *ast.Call {
	// In our aster script language, a call does not have
	// parens. It's like shell script, so if we have
	// an identifier and then a literal after we can assume
	// its a call
	var args []ast.Value

loop:
	for {
		t, pos, lit := p.lexer.Next()

		if pos.Line != cmdPos.Line || t == token.EOF {
			// TODO: unread token
			break
		}

		var val ast.Value
		switch t {
			case token.RBRACKET:
				p.lexer.Back()
				break loop
			case token.LPAREN:
				// TODO: Parse bracket statements
				continue
			case token.RPAREN:
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

	return &ast.Call{
		Name: cmd,
		Arguments: args,
	}
}

// Block parses a block.
func (p *Parser) Block() *ast.Block {
	_, lbrPos, _ := p.expectToken(token.LBRACKET)

	var list []ast.Node
	for {
		t, pos, lit := p.next()
		switch t {
			case token.IDENT, token.VAR:
				list = append(list, p.Ident(lit, pos))
			case token.RBRACKET:
				return &ast.Block{
					LBracket: lbrPos,
					List: list,
					RBracket: pos,
				}
		}
	}
}

func (p *Parser) CommandDecl() *ast.CommandDeclaration {
	_, pos, lit := p.expectToken(token.IDENT) // name of command

	n := &ast.CommandDeclaration{
		Pos: pos,
		Name: lit,
	}

	p.expectToken(token.LPAREN)

	var params []string
	for {
		t3, _, param := p.next()
		if t3 == token.IDENT {
			params = append(params, param)
		} else if t3 == token.RPAREN {
			break
		} else {
			panic("????")
		}
	}
	fmt.Println(params)

	n.Body = p.Block()
	n.Signature = params

	return n
}

// Filter parses a filter declaration.
// This is in the form of:
// filter name { ... }
func (p *Parser) Filter() *ast.FilterDeclaration {
	_, pos, lit := p.expectToken(token.IDENT) // name of filter

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
		panic(fmt.Errorf("unexpected EOF"))
	}

	return t, pos, lit
}

func (p *Parser) expectToken(expected token.Token) (token.Token, token.Position, string) {
	t, pos, lit := p.next()

	if t != expected {
		panic(fmt.Errorf("expected %s, found %s", expected, t))
	}

	return t, pos, lit
}
