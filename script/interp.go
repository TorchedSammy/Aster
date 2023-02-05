package script

import (
	"fmt"
	"io"
)

type Interpreter struct{
	vars map[string]*Value
	funs map[string]*Fun
}

// An aster function
type Fun struct{
	Caller func([]Value) []Value 
}

func NewInterp() *Interpreter {
	intr := &Interpreter{
		vars: make(map[string]*Value),
		funs: make(map[string]*Fun),
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
	i.funs[name] = &f
}

func (i *Interpreter) Run(r io.Reader) {
	nodes, err := Parse(r)
	if err != nil {
		panic(err)
	}

	for _, node := range nodes {
		switch n := node.(type) {
			case Decl:
				fmt.Printf("!! assigning %s to a value of \"%s\"\n", n.Name, n.Val)
				i.vars[n.Name] = &n.Val
			case Call:
				if i.funs[n.Name] == nil {
					panic("unknown function " + n.Name)
				}

				args := []Value{}

				for _, arg := range n.Arguments {
					if arg.Kind == VariableKind {
						if i.vars[arg.Val] == nil {
							panic(fmt.Sprintf("undefined variable %s", arg.Val))
						}

						args = append(args, *i.vars[arg.Val])
					}

					args = append(args, arg)
				}

				i.funs[n.Name].Caller(args)
		}
	}
}
