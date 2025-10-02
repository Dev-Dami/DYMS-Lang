package ast

import (
	"bytes"
	"fmt"
	"strings"
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
	IfStatementNode    NodeType = "IfStatement"
	ForStatementNode   NodeType = "ForStatement"
	WhileStatementNode NodeType = "WhileStatement"
	BlockStatementNode NodeType = "BlockStatement"
	AssignmentExprNode NodeType = "AssignmentExpr"
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

// Block Statement

type BlockStatement struct {
	Statements []Stmt
}

func (bs *BlockStatement) Kind() NodeType { return BlockStatementNode }

// If Statement

type IfStatement struct {
	Condition   Expr
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) Kind() NodeType { return IfStatementNode }

// For Statement

type ForStatement struct {
	Identifier *Identifier
	Range      Expr
	Body       *BlockStatement
}

func (fs *ForStatement) Kind() NodeType { return ForStatementNode }

// While Statement

type WhileStatement struct {
	Condition Expr
	Body      *BlockStatement
}

func (ws *WhileStatement) Kind() NodeType { return WhileStatementNode }

// Assignment Expression

type AssignmentExpr struct {
	Assignee Expr
	Value    Expr
}

func (a *AssignmentExpr) Kind() NodeType { return AssignmentExprNode }
func (a *AssignmentExpr) exprNode()      {}

func PrettyPrint(e Stmt) string {
	switch node := e.(type) {
	case *NumericLiteral:
		return fmt.Sprintf("%v", node.Value)
	case *Identifier:
		return node.Symbol
	case *StringLiteral:
		return fmt.Sprintf("\"%s\"", node.Value)
	case *BinaryExpr:
		left := PrettyPrint(node.Left)
		right := PrettyPrint(node.Right)
		return fmt.Sprintf("(%s %s %s)", left, node.Operator, right)
	case *IfStatement:
		var out bytes.Buffer
		out.WriteString("if ")
		out.WriteString(PrettyPrint(node.Condition))
		out.WriteString(" ")
		out.WriteString(PrettyPrint(node.Consequence))
		if node.Alternative != nil {
			out.WriteString(" else ")
			out.WriteString(PrettyPrint(node.Alternative))
		}
		return out.String()
	case *BlockStatement:
		var out bytes.Buffer
		out.WriteString("{\n")
		for _, s := range node.Statements {
			out.WriteString(PrettyPrint(s))
			out.WriteString("\n")
		}
		out.WriteString("}")
		return out.String()
	case *Program:
		var out bytes.Buffer
		for _, s := range node.Body {
			out.WriteString(PrettyPrint(s))
		}
		return out.String()
	case *VarDeclaration:
		var out bytes.Buffer
		if node.Constant {
			out.WriteString("const ")
		} else {
			out.WriteString("let ")
		}
		out.WriteString(node.Identifier)
		out.WriteString(" = ")
		out.WriteString(PrettyPrint(node.Value))
		return out.String()
	case *ForStatement:
		var out bytes.Buffer
		out.WriteString("for ")
		out.WriteString(PrettyPrint(node.Identifier))
		out.WriteString(" in ")
		out.WriteString(PrettyPrint(node.Range))
		out.WriteString(" ")
		out.WriteString(PrettyPrint(node.Body))
		return out.String()
	case *WhileStatement:
		var out bytes.Buffer
		out.WriteString("while ")
		out.WriteString(PrettyPrint(node.Condition))
		out.WriteString(" ")
		out.WriteString(PrettyPrint(node.Body))
		return out.String()
	case *CallExpr:
		var out bytes.Buffer
		args := []string{}
		for _, arg := range node.Args {
			args = append(args, PrettyPrint(arg))
		}
		out.WriteString(PrettyPrint(node.Callee))
		out.WriteString("(")
		out.WriteString(strings.Join(args, ", "))
		out.WriteString(")")
		return out.String()
	case *AssignmentExpr:
		var out bytes.Buffer
		out.WriteString(PrettyPrint(node.Assignee))
		out.WriteString(" = ")
		out.WriteString(PrettyPrint(node.Value))
		return out.String()
	default:
		return fmt.Sprintf("Unknown statement type: %T", e)
	}
}

