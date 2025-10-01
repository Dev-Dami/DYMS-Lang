package parser

import (
	"fmt"
	"holygo/ast"
	"holygo/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

// Create new parser
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Helpers
func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: -1, Value: ""}
	}
	return p.tokens[p.pos]
}

func (p *Parser) consume() lexer.Token {
	tok := p.peek()
	p.pos++
	return tok
}

// Parse entire program
func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{Body: []ast.Stmt{}}
	for p.pos < len(p.tokens) {
		stmt := p.parseStmt()
		if stmt != nil {
			prog.Body = append(prog.Body, stmt)
		}
	}
	return prog
}

// For now, treat everything as an expression statement
func (p *Parser) parseStmt() ast.Stmt {
	return p.parseExpr()
}

// Parse an expression (supports binary +, -, *, /)
func (p *Parser) parseExpr() ast.Expr {
	left := p.parsePrimary()

	for {
		tok := p.peek()
		if tok.Type == lexer.BinaryOperator {
			op := p.consume().Value
			right := p.parsePrimary()
			left = &ast.BinaryExpr{
				Left:     left,
				Right:    right,
				Operator: op,
			}
		} else {
			break
		}
	}
	return left
}

// Parse basic literals and identifiers
func (p *Parser) parsePrimary() ast.Expr {
	tok := p.consume()
	switch tok.Type {
	case lexer.Number:
		return &ast.NumericLiteral{Value: toNumber(tok.Value)}
	case lexer.Identifier:
		return &ast.Identifier{Symbol: tok.Value}
	case lexer.OpenParen:
		expr := p.parseExpr()
		if p.peek().Type == lexer.CloseParen {
			p.consume()
		}
		return expr
	default:
		fmt.Printf("Unexpected token: %s\n", tok.Value)
		return nil
	}
}

// Helper to convert string -> float64
func toNumber(s string) float64 {
	var n float64
	fmt.Sscanf(s, "%f", &n)
	return n
}
