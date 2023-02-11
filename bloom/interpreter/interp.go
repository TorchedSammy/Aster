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
}

type Interpreter struct{
	s *Scope
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
	nodes, err := p.Parse(r)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		switch n := node.(type) {
			case *ast.Decl:
				//fmt.Printf("!! assigning %s to a value of \"%s\"\n", n.Name, n.Val)
				i.s.Vars[n.Name] = n.Val
			case *ast.Call:
				if i.s.Funs[n.Name] == nil {
					return fmt.Errorf(interpErrorMessages[UndefinedCommand], n.Name)
				}

				args := []ast.Value{}

				for _, arg := range n.Arguments {
					if arg.Kind == ast.VariableKind {
						if i.s.Vars[arg.Val] == ast.EmptyValue {
							return fmt.Errorf(interpErrorMessages[UndefinedVariable], arg.Val)
						}

						args = append(args, i.s.Vars[arg.Val])
					}

					args = append(args, arg)
				}

				i.s.Funs[n.Name].Caller(args)
			case *ast.FilterDeclaration:
				fmt.Printf("declaring a new filter with name %s\n", n.Name)
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
