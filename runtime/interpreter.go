package runtime

import (
	"fmt"
	"strconv"
	"holygo/lexer"
)

// Node types
type NodeType int

const (
	NumericLiteral NodeType = iota
	Identifier
	BinaryExprNode 
)

// AST Nodes
type Stmt interface {
	Type() NodeType
}

type Expr interface {
	Stmt
	Value() interface{}
}

// Expressions
type NumericLiteralExpr struct {
	Val float64
}

func (n *NumericLiteralExpr) Type() NodeType {
	return NumericLiteral
}

func (n *NumericLiteralExpr) Value() interface{} {
	return n.Val
}

type IdentifierExpr struct {
	Symbol string
}

func (i *IdentifierExpr) Type() NodeType {
	return Identifier
}

func (i *IdentifierExpr) Value() interface{} {
	return i.Symbol
}

type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator string
}

func (b *BinaryExpr) Type() NodeType {
	return BinaryExprNode // updated to match constant
}

func (b *BinaryExpr) Value() interface{} {
	// In a real interpreter, you'd evaluate the expression here.
	return fmt.Sprintf("(%v %s %v)", b.Left.Value(), b.Operator, b.Right.Value())
}

// Parser
type Parser struct {
	tokens []lexer.Token
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (Stmt, error) {
	return p.parseExpr()
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseAdditiveExpr()
}

func (p *Parser) parseAdditiveExpr() (Expr, error) {
	left, err := p.parseMultiplicativeExpr()
	if err != nil {
		return nil, err
	}

	for len(p.tokens) > 0 && (p.tokens[0].Value == "+" || p.tokens[0].Value == "-") {
		operator := p.tokens[0].Value
		p.tokens = p.tokens[1:] // Consume operator

		right, err := p.parseMultiplicativeExpr()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}

	return left, nil
}

func (p *Parser) parseMultiplicativeExpr() (Expr, error) {
	left, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}

	for len(p.tokens) > 0 && (p.tokens[0].Value == "*" || p.tokens[0].Value == "/") {
		operator := p.tokens[0].Value
		p.tokens = p.tokens[1:] // Consume operator

		right, err := p.parsePrimaryExpr()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{Left: left, Right: right, Operator: operator}
	}

	return left, nil
}

func (p *Parser) parsePrimaryExpr() (Expr, error) {
	if len(p.tokens) == 0 {
		return nil, fmt.Errorf("unexpected end of input")
	}

	token := p.tokens[0]
	p.tokens = p.tokens[1:]

	switch token.Type {
	case lexer.Number:
		val, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %v", token.Value)
		}
		return &NumericLiteralExpr{Val: val}, nil
	case lexer.Identifier:
		return &IdentifierExpr{Symbol: token.Value}, nil
	case lexer.OpenParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if len(p.tokens) == 0 || p.tokens[0].Type != lexer.CloseParen {
			return nil, fmt.Errorf("expected ')'")
		}
		p.tokens = p.tokens[1:] // Consume ')'
		return expr, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
	}
}

// Evaluator
func Evaluate(stmt Stmt) (interface{}, error) {
	switch s := stmt.(type) {
	case *NumericLiteralExpr:
		return s.Val, nil
	case *IdentifierExpr:
		// In a real interpreter, you'd look up the variable's value.

		return s.Symbol, nil
	case *BinaryExpr:
		leftVal, err := Evaluate(s.Left)
		if err != nil {
			return nil, err
		}
		rightVal, err := Evaluate(s.Right)
		if err != nil {
			return nil, err
		}

		leftFloat, okLeft := leftVal.(float64)
		rightFloat, okRight := rightVal.(float64)

		if !okLeft || !okRight {
			return nil, fmt.Errorf("binary operations can only be performed on numbers")
		}

		switch s.Operator {
		case "+":
			return leftFloat + rightFloat, nil
		case "-":
			return leftFloat - rightFloat, nil
		case "*":
			return leftFloat * rightFloat, nil
		case "/":
			if rightFloat == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return leftFloat / rightFloat, nil
		default:
			return nil, fmt.Errorf("unknown operator: %s", s.Operator)
		}
	default:
		return nil, fmt.Errorf("unknown statement type: %T", s)
	}
}
