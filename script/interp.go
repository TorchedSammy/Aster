package script

import (
	"fmt"
	"io"
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
	Vars map[string]Value
	Funs map[string]*Fun
}

type Interpreter struct{
	s *Scope
}

// An aster function
type Fun struct{
	Caller func([]Value) []Value 
}

func NewInterp() *Interpreter {
	intr := &Interpreter{
		s: NewScope(nil),
	}

	intr.RegisterFunction("print", Fun{
		Caller: func(v []Value) []Value {
			if v[0] == EmptyValue {
				return []Value{}
			}

			if v[0].Kind != StringKind {
				panic("expected string and did not get it")
			}

			fmt.Println(v[0].Val)

			return []Value{}
		},
	})

	return intr
}

func (i *Interpreter) RegisterFunction(name string, f Fun) {
	i.s.Funs[name] = &f
}

func (i *Interpreter) Run(r io.Reader) error {
	nodes, err := Parse(r)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		switch n := node.(type) {
			case Decl:
				fmt.Printf("!! assigning %s to a value of \"%s\"\n", n.Name, n.Val)
				i.s.Vars[n.Name] = n.Val
			case Call:
				if i.s.Funs[n.Name] == nil {
					return fmt.Errorf(interpErrorMessages[UndefinedCommand], n.Name)
				}

				args := []Value{}

				for _, arg := range n.Arguments {
					if arg.Kind == VariableKind {
						if i.s.Vars[arg.Val] == EmptyValue {
							return fmt.Errorf(interpErrorMessages[UndefinedVariable], arg.Val)
						}

						args = append(args, i.s.Vars[arg.Val])
					}

					args = append(args, arg)
				}

				i.s.Funs[n.Name].Caller(args)
		}
	}

	return nil
}

func (i *Interpreter) GetGlobal(name string) Value {
	return i.s.Vars[name]
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		Outer: outer,
		Vars: make(map[string]Value),
		Funs: make(map[string]*Fun),
	}
}
