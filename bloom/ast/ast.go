package ast

import (
	"github.com/TorchedSammy/Aster/bloom/token"
)

type Node interface {
	Start() token.Position
	End() token.Position
}

type Decl struct{
	Name string
	Pos token.Position
	Val Value
}

func (d Decl) Start() token.Position {
	return d.Pos
}

func (d Decl) End() token.Position {
	return d.Val.End()
}

type BadDecl struct {
	Name string
}

type Expr struct{
	Pos token.Position
}

func (e Expr) Start() token.Position {
	return e.Pos
}

func (e Expr) End() token.Position {
	return e.Pos
}

type ParenExpr struct{
	LParen token.Position
	E Expr
	RParen token.Position
}

func (p ParenExpr) Start() token.Position {
	return p.LParen
}

func (p ParenExpr) End() token.Position {
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
	Pos token.Position
	Val string
	Kind ValueKind
}
var EmptyValue = Value{}

func (v Value) Start() token.Position {
	return v.Pos
}

func (v Value) End() token.Position {
	return v.Pos
}

func (v Value) String() string {
	return v.Val
}

type Call struct {
	Pos token.Position
	Name string // name of function
	Arguments []Value
}

func (c Call) Start() token.Position {
	return c.Pos
}

func (c Call) End() token.Position {
	return c.Pos
}

type SwitchAssign struct {
	Pos token.Position
	Name string
	Val Value
	IsToggle bool
}

func (s SwitchAssign) Start() token.Position {
	return s.Pos
}

func (s SwitchAssign) End() token.Position {
	return s.Pos
}

type Block struct {
	LBracket token.Position
	List []Node
	RBracket token.Position
	InFilter bool
}

type FilterType int
const (
	ColorFilter FilterType = iota
)

type FilterDeclaration struct {
	Pos token.Position
	Name string
	Body *Block
}

func (f FilterDeclaration) Start() token.Position {
	return f.Pos
}

func (f FilterDeclaration) End() token.Position {
	return f.Pos
}
