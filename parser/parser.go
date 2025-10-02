package parser

import (
	"fmt"
	"holygo/ast"
	"holygo/lexer"
	"strconv"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

// create new parser
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// helpers
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

func (p *Parser) expect(expected lexer.TokenType, message string) lexer.Token {
	tok := p.consume()
	if tok.Type != expected {
		panic(fmt.Sprintf("%s at line %d, column %d", message, tok.Line, tok.Column))
	}
	return tok
}

// parse entire program
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

// treat everything -> expression statement
func (p *Parser) parseStmt() ast.Stmt {
	switch p.peek().Type {
	case lexer.Let, lexer.Var, lexer.Const:
		return p.parseVarDeclaration()
	case lexer.If:
		return p.parseIfStatement()
	case lexer.ForRange:
		return p.parseForStatement()
	case lexer.While:
		return p.parseWhileStatement()
	case lexer.Else:
		return nil // Ignore else tokens, as they are handled by parseIfStatement
	case lexer.OpenBrace:
		return p.parseBlockStatement()
	default:
		return p.parseExpr()
	}
}

func (p *Parser) parseVarDeclaration() ast.Stmt {
	isConstant := p.consume().Type == lexer.Const
	identifier := p.expect(lexer.Identifier, "Expected identifier in variable declaration").Value

	p.expect(lexer.Equals, "Expected '=' after identifier in variable declaration")
	value := p.parseExpr()
	return &ast.VarDeclaration{Identifier: identifier, Value: value, Constant: isConstant}
}

func (p *Parser) parseIfStatement() ast.Stmt {
	p.consume() // consume 'if'
	p.expect(lexer.OpenParen, "Expected '(' after 'if'")
	condition := p.parseExpr()
	p.expect(lexer.CloseParen, "Expected ')' after if condition")
	consequence := p.parseBlockStatement()

	var alternative *ast.BlockStatement
	if p.peek().Type == lexer.Else {
		p.consume() // consume 'else'
		alternative = p.parseBlockStatement()
	}

	return &ast.IfStatement{
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternative,
	}
}

func (p *Parser) parseForStatement() ast.Stmt {
	p.consume() // consume 'for range'
	p.expect(lexer.OpenParen, "Expected '(' after 'for range'")
	identifier := p.expect(lexer.Identifier, "Expected identifier in for loop").Value
	p.expect(lexer.Comma, "Expected ',' after identifier in for loop")
	rangeExpr := p.parseExpr()
	p.expect(lexer.CloseParen, "Expected ')' after for loop range")
	body := p.parseBlockStatement()

	return &ast.ForStatement{
		Identifier: &ast.Identifier{Symbol: identifier},
		Range:      rangeExpr,
		Body:       body,
	}
}

func (p *Parser) parseWhileStatement() ast.Stmt {
	p.consume() // consume 'while'
	p.expect(lexer.OpenParen, "Expected '(' after 'while'")
	condition := p.parseExpr()
	p.expect(lexer.CloseParen, "Expected ')' after while condition")
	body := p.parseBlockStatement()

	return &ast.WhileStatement{
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	p.expect(lexer.OpenBrace, "Expected '{' to start a block statement")
	statements := []ast.Stmt{}
	for p.peek().Type != lexer.CloseBrace && p.pos < len(p.tokens) {
		statements = append(statements, p.parseStmt())
	}
	p.expect(lexer.CloseBrace, "Expected '}' to end a block statement")
	return &ast.BlockStatement{Statements: statements}
}

// parse an expression
func (p *Parser) parseExpr() ast.Expr {
	return p.parseAssignmentExpr()
}

func (p *Parser) parseAssignmentExpr() ast.Expr {
	left := p.parseLogicalExpr()

	if p.peek().Type == lexer.Equals {
		fmt.Printf("Assignee in parseAssignmentExpr: %T\n", left)
		_, isIdentifier := left.(*ast.Identifier)
		if !isIdentifier {
			panic(fmt.Sprintf("Invalid assignment target: %T at line %d, column %d", left, p.peek().Line, p.peek().Column))
		}
		p.consume() // consume '='
		value := p.parseAssignmentExpr()
		return &ast.AssignmentExpr{Assignee: left, Value: value}
	}

	return left
}

func (p *Parser) parseLogicalExpr() ast.Expr {
	left := p.parseComparisonExpr()

	for p.peek().Type == lexer.LogicalOperator {
		op := p.consume().Value
		right := p.parseComparisonExpr()
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left
}

func (p *Parser) parseComparisonExpr() ast.Expr {
	left := p.parseAdditiveExpr()

	for p.peek().Type == lexer.ComparisonOperator {
		op := p.consume().Value
		right := p.parseAdditiveExpr()
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left
}

func (p *Parser) parseAdditiveExpr() ast.Expr {
	left := p.parseMultiplicativeExpr()

	for p.peek().Value == "+" || p.peek().Value == "-" {
		op := p.consume().Value
		right := p.parseMultiplicativeExpr()
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left
}

func (p *Parser) parseMultiplicativeExpr() ast.Expr {
	left := p.parseCallExpr()

	for p.peek().Value == "*" || p.peek().Value == "/" {
		op := p.consume().Value
		right := p.parseCallExpr()
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left
}

func (p *Parser) parseCallExpr() ast.Expr {
	callee := p.parsePrimary()

	if p.peek().Type == lexer.OpenParen {
		p.consume() // consume open paren
		args := []ast.Expr{}
		if p.peek().Type != lexer.CloseParen {
			for {
				args = append(args, p.parseExpr())
				if p.peek().Type == lexer.CloseParen {
					break
				}
				p.expect(lexer.Comma, fmt.Sprintf("Expected ',' or ')' in argument list, but got %s", p.peek().Value))
			}
		}
		p.expect(lexer.CloseParen, "Expected ')' after arguments")
		return &ast.CallExpr{Callee: callee, Args: args}
	}

	return callee
}

// parse literals and identifiers
func (p *Parser) parsePrimary() ast.Expr {
	tok := p.consume()
	switch tok.Type {
	case lexer.Number:
		return &ast.NumericLiteral{Value: toNumber(tok.Value)}
	case lexer.Identifier:
		return &ast.Identifier{Symbol: tok.Value}
	case lexer.String:
		return &ast.StringLiteral{Value: tok.Value}
	case lexer.OpenParen:
		expr := p.parseExpr()
		p.expect(lexer.CloseParen, "Expected ')' after expression in parentheses")
		return expr
	default:
		panic(fmt.Sprintf("Unexpected token: %s at line %d, column %d", tok.Value, tok.Line, tok.Column))
	}
}

// convert string -> float64
func toNumber(s string) float64 {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Sprintf("Could not parse number: %s", s))
	}
	return n
}
