package ast

import (
	"fmt"
)

type NodeType string

const (
	ProgramNode       NodeType = "Program"
	NumericLiteralNode NodeType = "NumericLiteral"
	IdentifierNode     NodeType = "Identifier"
	BinaryExprNode     NodeType = "BinaryExpr"
	VarDeclarationNode NodeType = "VarDeclaration"
	CallExprNode       NodeType = "CallExpr"
	StringLiteralNode  NodeType = "StringLiteral"
)

//stat interface
type Stmt interface {
	Kind() NodeType
}


// expr is a statement
type Expr interface {
	Stmt
	exprNode()
}

//program block statement
type Program struct {
	Body []Stmt
}

func (p *Program) Kind() NodeType { return ProgramNode }

//Binary EXPR operations

type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator string
}

func (b *BinaryExpr) Kind() NodeType { return BinaryExprNode }
func (b *BinaryExpr) exprNode()      {}


//Indentifier -> usser def variable / symbol

type Identifier struct {
	Symbol string
}

func (i *Identifier) Kind() NodeType { return IdentifierNode }
func (i *Identifier) exprNode()      {}


//Numerical Representations

type NumericLiteral struct {
	Value float64
}

func (n *NumericLiteral) Kind() NodeType { return NumericLiteralNode }
func (n *NumericLiteral) exprNode()      {}

// String Literal
type StringLiteral struct {
	Value string
}

func (s *StringLiteral) Kind() NodeType { return StringLiteralNode }
func (s *StringLiteral) exprNode()      {}

// Variable Declaration
type VarDeclaration struct {
	Identifier string
	Value      Expr
	Constant   bool
}

func (v *VarDeclaration) Kind() NodeType { return VarDeclarationNode }
func (v *VarDeclaration) exprNode()      {}

// Call Expression
type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (c *CallExpr) Kind() NodeType { return CallExprNode }
func (c *CallExpr) exprNode()      {}

func PrettyPrint(e Expr) string {
	switch node := e.(type) {
	case *NumericLiteral:
		return fmt.Sprintf("%v", node.Value)
	case *Identifier:
		return node.Symbol
	case *BinaryExpr:
		left := PrettyPrint(node.Left)
		right := PrettyPrint(node.Right)
		return fmt.Sprintf("(%s %s %s)", left, node.Operator, right)
	default:
		return "UnknownExpr"
	}
}

