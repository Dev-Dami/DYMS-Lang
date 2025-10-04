package ast

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type NodeType string

const (
	ProgramNode        NodeType = "Program"
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
	BooleanLiteralNode NodeType = "BooleanLiteral"
	ArrayLiteralNode   NodeType = "ArrayLiteral"
	MapLiteralNode     NodeType = "MapLiteral"
	MemberExprNode     NodeType = "MemberExpr"
	ImportStatementNode NodeType = "ImportStatement"
	FunctionDeclarationNode NodeType = "FunctionDeclaration"
	ReturnStatementNode    NodeType = "ReturnStatement"
	UnaryExprNode          NodeType = "UnaryExpr"
	TryStatementNode       NodeType = "TryStatement"
	CatchStatementNode     NodeType = "CatchStatement"
)

type Stmt interface {
	Kind() NodeType
}

type Expr interface {
	Stmt
	exprNode()
}

type Program struct {
	Body []Stmt
}
func (p *Program) Kind() NodeType { return ProgramNode }

type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator string
}
func (b *BinaryExpr) Kind() NodeType { return BinaryExprNode }
func (b *BinaryExpr) exprNode()      {}

type Identifier struct {
	Symbol string
}
func (i *Identifier) Kind() NodeType { return IdentifierNode }
func (i *Identifier) exprNode()      {}

type NumericLiteral struct {
	Value float64
}
func (n *NumericLiteral) Kind() NodeType { return NumericLiteralNode }
func (n *NumericLiteral) exprNode()      {}

type StringLiteral struct {
	Value string
}
func (s *StringLiteral) Kind() NodeType { return StringLiteralNode }
func (s *StringLiteral) exprNode()      {}

type VarDeclaration struct {
	Identifier string
	Value      Expr
	Constant   bool
}
func (v *VarDeclaration) Kind() NodeType { return VarDeclarationNode }
func (v *VarDeclaration) exprNode()      {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
}
func (c *CallExpr) Kind() NodeType { return CallExprNode }
func (c *CallExpr) exprNode()      {}

type MemberExpr struct {
	Object   Expr
	Property *Identifier
}
func (m *MemberExpr) Kind() NodeType { return MemberExprNode }
func (m *MemberExpr) exprNode()      {}

type BlockStatement struct {
	Statements []Stmt
}
func (bs *BlockStatement) Kind() NodeType { return BlockStatementNode }

type IfStatement struct {
	Condition   Expr
	Consequence *BlockStatement
	Alternative *BlockStatement
}
func (is *IfStatement) Kind() NodeType { return IfStatementNode }

type ForStatement struct {
	Identifier *Identifier
	Range      Expr
	Body       *BlockStatement
}
func (fs *ForStatement) Kind() NodeType { return ForStatementNode }

type WhileStatement struct {
	Condition Expr
	Body      *BlockStatement
}
func (ws *WhileStatement) Kind() NodeType { return WhileStatementNode }

type AssignmentExpr struct {
	Assignee Expr
	Value    Expr
}
func (a *AssignmentExpr) Kind() NodeType { return AssignmentExprNode }
func (a *AssignmentExpr) exprNode()      {}

type BooleanLiteral struct {
	Value bool
}
func (b *BooleanLiteral) Kind() NodeType { return BooleanLiteralNode }
func (b *BooleanLiteral) exprNode()      {}

type ArrayLiteral struct {
	Elements []Expr
}
func (a *ArrayLiteral) Kind() NodeType { return ArrayLiteralNode }
func (a *ArrayLiteral) exprNode()      {}

type MapLiteral struct {
	Properties []*Property
}
func (m *MapLiteral) Kind() NodeType { return MapLiteralNode }
func (m *MapLiteral) exprNode()      {}

type ImportStatement struct {
	Path  string
	Alias string
}
func (is *ImportStatement) Kind() NodeType { return ImportStatementNode }

// Function Declaration: funct name(a, b) { ... }
type FunctionDeclaration struct {
	Name   string
	Params []string
	Body   *BlockStatement
}
func (fd *FunctionDeclaration) Kind() NodeType { return FunctionDeclarationNode }

// Return Statement: return expr
type ReturnStatement struct {
	Value Expr
}
func (rs *ReturnStatement) Kind() NodeType { return ReturnStatementNode }

// Unary expressions: ++x, --x, x++, x--
type UnaryExpr struct {
	Operand  Expr
	Operator string
	Prefix   bool // true for ++x, false for x++
}
func (u *UnaryExpr) Kind() NodeType { return UnaryExprNode }
func (u *UnaryExpr) exprNode()      {}

// Try-catch statement: try { ... } catch(e) { ... }
type TryStatement struct {
	TryBlock   *BlockStatement
	CatchBlock *BlockStatement
	ErrorVar   string
}
func (ts *TryStatement) Kind() NodeType { return TryStatementNode }

type Property struct {
	Key   Expr
	Value Expr
}

func PrettyPrint(e Stmt) string {
	start := time.Now()
	var result string
	switch node := e.(type) {
	case *NumericLiteral:
		result = fmt.Sprintf("%v", node.Value)
	case *Identifier:
		result = node.Symbol
	case *StringLiteral:
		result = fmt.Sprintf("\"%s\"", node.Value)
	case *BinaryExpr:
		left := PrettyPrint(node.Left)
		right := PrettyPrint(node.Right)
		result = fmt.Sprintf("(%s %s %s)", left, node.Operator, right)
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
		result = out.String()
	case *BlockStatement:
		var out bytes.Buffer
		out.WriteString("{\n")
		for _, s := range node.Statements {
			out.WriteString(PrettyPrint(s))
			out.WriteString("\n")
		}
		out.WriteString("}")
		result = out.String()
	case *Program:
		var out bytes.Buffer
		for _, s := range node.Body {
			out.WriteString(PrettyPrint(s))
		}
		result = out.String()
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
		result = out.String()
	case *ForStatement:
		var out bytes.Buffer
		out.WriteString("for ")
		out.WriteString(PrettyPrint(node.Identifier))
		out.WriteString(" in ")
		out.WriteString(PrettyPrint(node.Range))
		out.WriteString(" ")
		out.WriteString(PrettyPrint(node.Body))
		result = out.String()
	case *WhileStatement:
		var out bytes.Buffer
		out.WriteString("while ")
		out.WriteString(PrettyPrint(node.Condition))
		out.WriteString(" ")
		out.WriteString(PrettyPrint(node.Body))
		result = out.String()
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
		result = out.String()
	case *MemberExpr:
		result = fmt.Sprintf("%s.%s", PrettyPrint(node.Object), node.Property.Symbol)
	case *ImportStatement:
		result = fmt.Sprintf("import \"%s\" as %s", node.Path, node.Alias)
	case *AssignmentExpr:
		var out bytes.Buffer
		out.WriteString(PrettyPrint(node.Assignee))
		out.WriteString(" = ")
		out.WriteString(PrettyPrint(node.Value))
		result = out.String()
	case *BooleanLiteral:
		result = fmt.Sprintf("%v", node.Value)
	case *ArrayLiteral:
		var out bytes.Buffer
		var elements []string
		for _, el := range node.Elements {
			elements = append(elements, PrettyPrint(el))
		}
		out.WriteString("[")
		out.WriteString(strings.Join(elements, ", "))
		out.WriteString("]")
		result = out.String()
case *MapLiteral:
		var out bytes.Buffer
		var properties []string
		for _, prop := range node.Properties {
			properties = append(properties, fmt.Sprintf("%s: %s", PrettyPrint(prop.Key), PrettyPrint(prop.Value)))
		}
		out.WriteString("{")
		out.WriteString(strings.Join(properties, ", "))
		out.WriteString("}")
		result = out.String()
	case *FunctionDeclaration:
		var out bytes.Buffer
		out.WriteString("funct ")
		out.WriteString(node.Name)
		out.WriteString("(")
		out.WriteString(strings.Join(node.Params, ", "))
		out.WriteString(") ")
		out.WriteString(PrettyPrint(node.Body))
		result = out.String()
	case *ReturnStatement:
		result = fmt.Sprintf("return %s", PrettyPrint(node.Value))
	case *UnaryExpr:
		if node.Prefix {
			result = fmt.Sprintf("%s%s", node.Operator, PrettyPrint(node.Operand))
		} else {
			result = fmt.Sprintf("%s%s", PrettyPrint(node.Operand), node.Operator)
		}
	case *TryStatement:
		var out bytes.Buffer
		out.WriteString("try ")
		out.WriteString(PrettyPrint(node.TryBlock))
		out.WriteString(" catch(")
		out.WriteString(node.ErrorVar)
		out.WriteString(") ")
		out.WriteString(PrettyPrint(node.CatchBlock))
		result = out.String()
	default:
		result = fmt.Sprintf("Unknown statement type: %T", e)
	}
	elapsed := time.Since(start)
	fmt.Printf("PrettyPrint took %s\n", elapsed)
	return result
}
