package parser

import (
	"fmt"
	"DYMS/ast"
	"DYMS/lexer"
	"DYMS/runtime"
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

func (p *Parser) expect(expected lexer.TokenType, message string) (lexer.Token, *runtime.Error) {
	tok := p.consume()
	if tok.Type != expected {
		return tok, runtime.NewError(fmt.Sprintf("%s at line %d, column %d", message, tok.Line, tok.Column), tok.Line, tok.Column)
	}
	return tok, nil
}

// parse entire program
func (p *Parser) ParseProgram() (*ast.Program, *runtime.Error) {
	prog := &ast.Program{Body: []ast.Stmt{}}
	for p.pos < len(p.tokens) {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Body = append(prog.Body, stmt)
		}
	}
	return prog, nil
}

// treat everything -> expression statement
func (p *Parser) parseStmt() (ast.Stmt, *runtime.Error) {
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
		return nil, nil // Ignore else tokens, as they are handled by parseIfStatement
	case lexer.OpenBrace:
		return p.parseBlockStatement()
	default:
		return p.parseExpr()
	}
}

func (p *Parser) parseVarDeclaration() (ast.Stmt, *runtime.Error) {
	isConstant := p.consume().Type == lexer.Const
	identifier, err := p.expect(lexer.Identifier, "Expected identifier in variable declaration")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(lexer.Equals, "Expected '=' after identifier in variable declaration")
	if err != nil {
		return nil, err
	}
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.VarDeclaration{Identifier: identifier.Value, Value: value, Constant: isConstant}, nil
}

func (p *Parser) parseIfStatement() (ast.Stmt, *runtime.Error) {
	p.consume() // consume 'if'
	_, err := p.expect(lexer.OpenParen, "Expected '(' after 'if'")
	if err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(lexer.CloseParen, "Expected ')' after if condition")
	if err != nil {
		return nil, err
	}
	consequence, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	var alternative *ast.BlockStatement
	if p.peek().Type == lexer.Else {
		p.consume() // consume 'else'
		alternative, err = p.parseBlockStatement()
		if err != nil {
			return nil, err
		}
	}

	return &ast.IfStatement{
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternative,
	}, nil
}

func (p *Parser) parseForStatement() (ast.Stmt, *runtime.Error) {
	p.consume() // consume 'for range'
	_, err := p.expect(lexer.OpenParen, "Expected '(' after 'for range'")
	if err != nil {
		return nil, err
	}
	identifier, err := p.expect(lexer.Identifier, "Expected identifier in for loop")
	if err != nil {
		return nil, err
	}
	_, err = p.expect(lexer.Comma, "Expected ',' after identifier in for loop")
	if err != nil {
		return nil, err
	}
	rangeExpr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(lexer.CloseParen, "Expected ')' after for loop range")
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.ForStatement{
		Identifier: &ast.Identifier{Symbol: identifier.Value},
		Range:      rangeExpr,
		Body:       body,
	}, nil
}

func (p *Parser) parseWhileStatement() (ast.Stmt, *runtime.Error) {
	p.consume() // consume 'while'
	_, err := p.expect(lexer.OpenParen, "Expected '(' after 'while'")
	if err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(lexer.CloseParen, "Expected ')' after while condition")
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.WhileStatement{
		Condition: condition,
		Body:      body,
	}, nil
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, *runtime.Error) {
	_, err := p.expect(lexer.OpenBrace, "Expected '{' to start a block statement")
	if err != nil {
		return nil, err
	}
	statements := []ast.Stmt{}
	for p.peek().Type != lexer.CloseBrace && p.pos < len(p.tokens) {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}
	_, err = p.expect(lexer.CloseBrace, "Expected '}' to end a block statement")
	if err != nil {
		return nil, err
	}
	return &ast.BlockStatement{Statements: statements}, nil
}

// parse an expression
func (p *Parser) parseExpr() (ast.Expr, *runtime.Error) {
	return p.parseAssignmentExpr()
}

func (p *Parser) parseAssignmentExpr() (ast.Expr, *runtime.Error) {
	left, err := p.parseLogicalExpr()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == lexer.Equals {
		_, isIdentifier := left.(*ast.Identifier)
		if !isIdentifier {
			return nil, runtime.NewError(fmt.Sprintf("Invalid assignment target: %T at line %d, column %d", left, p.peek().Line, p.peek().Column), p.peek().Line, p.peek().Column)
		}
		p.consume() // consume '='
		value, err := p.parseAssignmentExpr()
		if err != nil {
			return nil, err
		}
		return &ast.AssignmentExpr{Assignee: left, Value: value}, nil
	}

	return left, nil
}

func (p *Parser) parseLogicalExpr() (ast.Expr, *runtime.Error) {
	left, err := p.parseComparisonExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == lexer.LogicalOperator {
		op := p.consume().Value
		right, err := p.parseComparisonExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left, nil
}

func (p *Parser) parseComparisonExpr() (ast.Expr, *runtime.Error) {
	left, err := p.parseAdditiveExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == lexer.ComparisonOperator {
		op := p.consume().Value
		right, err := p.parseAdditiveExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left, nil
}

func (p *Parser) parseAdditiveExpr() (ast.Expr, *runtime.Error) {
	left, err := p.parseMultiplicativeExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Value == "+" || p.peek().Value == "-" {
		op := p.consume().Value
		right, err := p.parseMultiplicativeExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left, nil
}

func (p *Parser) parseMultiplicativeExpr() (ast.Expr, *runtime.Error) {
	left, err := p.parseCallExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Value == "*" || p.peek().Value == "/" {
		op := p.consume().Value
		right, err := p.parseCallExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}

	return left, nil
}

func (p *Parser) parseCallExpr() (ast.Expr, *runtime.Error) {
	callee, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	if p.peek().Type == lexer.OpenParen {
		p.consume() // consume open paren
		args := []ast.Expr{}
		if p.peek().Type != lexer.CloseParen {
			for {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.peek().Type == lexer.CloseParen {
					break
				}
				_, err = p.expect(lexer.Comma, fmt.Sprintf("Expected ',' or ')' in argument list, but got %s", p.peek().Value))
				if err != nil {
					return nil, err
				}
			}
		}
		_, err = p.expect(lexer.CloseParen, "Expected ')' after arguments")
		if err != nil {
			return nil, err
		}
		return &ast.CallExpr{Callee: callee, Args: args}, nil
	}

	return callee, nil
}

// parse literals and identifiers
func (p *Parser) parsePrimary() (ast.Expr, *runtime.Error) {
	tok := p.consume()
	switch tok.Type {
	case lexer.Number:
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, runtime.NewError(fmt.Sprintf("Could not parse number: %s", tok.Value), tok.Line, tok.Column)
		}
		return &ast.NumericLiteral{Value: val}, nil
	case lexer.Identifier:
		return &ast.Identifier{Symbol: tok.Value}, nil
	case lexer.String:
		return &ast.StringLiteral{Value: tok.Value}, nil
	case lexer.True:
		return &ast.BooleanLiteral{Value: true}, nil
	case lexer.False:
		return &ast.BooleanLiteral{Value: false}, nil
	case lexer.OpenBracket:
		return p.parseArrayLiteral()
	case lexer.OpenBrace:
		return p.parseMapLiteral()
	case lexer.OpenParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(lexer.CloseParen, "Expected ')' after expression in parentheses")
		if err != nil {
			return nil, err
		}
		return expr, nil
	default:
		return nil, runtime.NewError(fmt.Sprintf("Unexpected token: %s", tok.Value), tok.Line, tok.Column)
	}
}

func (p *Parser) parseArrayLiteral() (ast.Expr, *runtime.Error) {
	elements := []ast.Expr{}
	// '[' has already been consumed by parsePrimary
	if p.peek().Type != lexer.CloseBracket {
		for {
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)
			if p.peek().Type == lexer.CloseBracket {
				break
			}
			_, err = p.expect(lexer.Comma, fmt.Sprintf("Expected ',' or ']' in array literal, but got %s", p.peek().Value))
			if err != nil {
				return nil, err
			}
		}
	}
	_, err := p.expect(lexer.CloseBracket, "Expected ']' to end an array literal")
	if err != nil {
		return nil, err
	}
	return &ast.ArrayLiteral{Elements: elements}, nil
}

func (p *Parser) parseMapLiteral() (ast.Expr, *runtime.Error) {
	properties := []*ast.Property{}
	// '{' has already been consumed by parsePrimary
	if p.peek().Type != lexer.CloseBrace {
		for {
			key, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(lexer.Colon, "Expected ':' after key in map literal")
			if err != nil {
				return nil, err
			}
			value, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			properties = append(properties, &ast.Property{Key: key, Value: value})
			if p.peek().Type == lexer.CloseBrace {
				break
			}
			_, err = p.expect(lexer.Comma, fmt.Sprintf("Expected ',' or '}' in map literal, but got %s", p.peek().Value))
			if err != nil {
				return nil, err
			}
		}
	}
	_, err := p.expect(lexer.CloseBrace, "Expected '}' to end a map literal")
	if err != nil {
		return nil, err
	}
	return &ast.MapLiteral{Properties: properties}, nil
}


