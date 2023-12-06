package interpreter

import (
	"fmt"
	"io"
	"os"

	"github.com/TorchedSammy/Aster/bloom/ast"
	"github.com/TorchedSammy/Aster/bloom/parser"
)

type InterpErrorType int
const (
	UndefinedCommand InterpErrorType = iota
	UndefinedVariable
)

var interpErrorMessages = map[InterpErrorType]string{
	UndefinedCommand: "attempt to run undefined command %s",
	UndefinedVariable: "reference to undefined variable %s",
}

type Scope struct{
	Outer *Scope
	Vars map[string]ast.Value
	Funs map[string]*Fun
	Filters map[string]*Filter
}

type Interpreter struct{
	s *Scope
	fh *FilterHandler
}

// An aster function
type Fun struct{
	Caller func([]ast.Value) []ast.Value 
}

func NewInterp() *Interpreter {
	intr := &Interpreter{
		s: NewScope(nil),
	}

	intr.RegisterFunction("print", Fun{
		Caller: func(v []ast.Value) []ast.Value {
			if v[0] == ast.EmptyValue {
				return []ast.Value{}
			}

			if v[0].Kind != ast.StringKind {
				panic("expected string and did not get it")
			}

			fmt.Println(v[0].Val)

			return []ast.Value{}
		},
	})

	intr.RegisterFunction("source", Fun{
		Caller: func(v []ast.Value) []ast.Value {
			if v[0] == ast.EmptyValue {
				return []ast.Value{}
			}

			if v[0].Kind != ast.StringKind {
				panic("expected string and did not get it")
			}

			f, _ := os.Open(v[0].Val)
			intr.Run(f)

			return []ast.Value{}
		},
	})

	return intr
}

func (i *Interpreter) RegisterFunction(name string, f Fun) {
	i.s.Funs[name] = &f
}

func (i *Interpreter) Run(r io.Reader) error {
	p := parser.New()
	block, err := p.Parse(r)
	if err != nil {
		return err
	}

	return i.runBlock(block, i.s)
}

func (i *Interpreter) runBlock(b *ast.Block, s *Scope) error {
	for _, node := range b.List {
		switch n := node.(type) {
			case *ast.Decl:
				fmt.Printf("!! assigning %s to a value of \"%s\"\n", n.Name, n.Val)
				s.Vars[n.Name] = n.Val
			case *ast.Call:
				fmt.Println("running call", n.Name)
				if s.GetFunc(n.Name) == nil {
					return fmt.Errorf(interpErrorMessages[UndefinedCommand], n.Name)
				}

				args := []ast.Value{}

				for _, arg := range n.Arguments {
					if arg.Kind == ast.VariableKind {
						if s.GetVar(arg.Val) == ast.EmptyValue {
							return fmt.Errorf(interpErrorMessages[UndefinedVariable], arg.Val)
						}

						args = append(args, s.Vars[arg.Val])
					}

					args = append(args, arg)
				}

				i.s.Funs[n.Name].Caller(args)
			case *ast.FilterDeclaration:
				fmt.Printf("declaring a new filter with name %s\n", n.Name)
			case *ast.CommandDeclaration:
				i.RegisterFunction(n.Name, Fun{
					Caller: func([]ast.Value) []ast.Value {
						err := i.runBlock(n.Body, NewScope(s))
						if err != nil {
							fmt.Println(err)
						}

						return []ast.Value{}
					},
				})
		}
	}

	return nil
}

func (i *Interpreter) GetGlobal(name string) ast.Value {
	return i.s.Vars[name]
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		Outer: outer,
		Vars: make(map[string]ast.Value),
		Funs: make(map[string]*Fun),
	}
}

func (s *Scope) GetFunc(name string) *Fun {
	f := s.Funs[name]
	if f == nil && s.Outer != nil {
		f = s.Outer.GetFunc(name)
	}

	return f
}

func (s *Scope) GetVar(name string) ast.Value {
	v := s.Vars[name]
	if v == ast.EmptyValue && s.Outer != nil {
		v = s.Outer.GetVar(name)
	}

	return v
}
