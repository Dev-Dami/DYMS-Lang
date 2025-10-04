package parser

import (
	"fmt"
	"DYMS/ast"
	"DYMS/lexer"
	"DYMS/runtime"
	"strconv"
)

type Parser struct {
	tokens    []lexer.Token
	pos       int
	lookahead [3]lexer.Token // Fast lookahead cache
	lookaheadValid [3]bool
}

// parser ->
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// Fast helpers with caching
func (p *Parser) peek() lexer.Token {
	if !p.lookaheadValid[0] {
		if p.pos >= len(p.tokens) {
			p.lookahead[0] = lexer.Token{Type: -1, Value: ""}
		} else {
			p.lookahead[0] = p.tokens[p.pos]
		}
		p.lookaheadValid[0] = true
	}
	return p.lookahead[0]
}

func (p *Parser) peekAhead(offset int) lexer.Token {
	if offset >= 3 { // fallback for far lookahead
		if p.pos+offset >= len(p.tokens) {
			return lexer.Token{Type: -1, Value: ""}
		}
		return p.tokens[p.pos+offset]
	}
	if !p.lookaheadValid[offset] {
		if p.pos+offset >= len(p.tokens) {
			p.lookahead[offset] = lexer.Token{Type: -1, Value: ""}
		} else {
			p.lookahead[offset] = p.tokens[p.pos+offset]
		}
		p.lookaheadValid[offset] = true
	}
	return p.lookahead[offset]
}

func (p *Parser) consume() lexer.Token {
	tok := p.peek()
	p.pos++
	// Shift lookahead cache for better performance
	p.lookahead[0] = p.lookahead[1]
	p.lookahead[1] = p.lookahead[2]
	p.lookaheadValid[0] = p.lookaheadValid[1]
	p.lookaheadValid[1] = p.lookaheadValid[2]
	p.lookaheadValid[2] = false
	return tok
}

func (p *Parser) expect(expected lexer.TokenType, message string) (lexer.Token, *runtime.Error) {
	tok := p.consume()
	if tok.Type != expected {
		return tok, runtime.NewError(fmt.Sprintf("%s at line %d, column %d", message, tok.Line, tok.Column), tok.Line, tok.Column)
	}
	return tok, nil
}

// parseprogram ->
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

// parsestmt ->
func (p *Parser) parseStmt() (ast.Stmt, *runtime.Error) {
	switch p.peek().Type {
	case lexer.Import:
		return p.parseImportStatement()
	case lexer.Funct:
		return p.parseFunctionDeclaration()
	case lexer.Return:
		return p.parseReturnStatement()
	case lexer.Try:
		return p.parseTryStatement()
	case lexer.Break:
		p.consume()
		return &ast.BreakStatement{}, nil
	case lexer.Continue:
		p.consume()
		return &ast.ContinueStatement{}, nil
	case lexer.Let, lexer.Var, lexer.Const:
		return p.parseVarDeclaration()
	case lexer.If:
		return p.parseIfStatement()
	case lexer.ForRange:
		return p.parseForStatement()
	case lexer.While:
		return p.parseWhileStatement()
	case lexer.Else:
		return nil, nil // else -> handled by if
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
	p.consume() // if ->
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
		p.consume() // else ->
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
	p.consume() // for range ->
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
	p.consume() // while ->
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

// parseexpr ->
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
		p.consume() // = ->
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
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	// Fast operator checking
	for {
		tok := p.peek()
		if tok.Type == lexer.BinaryOperator {
			if tok.Value != "*" && tok.Value != "/" {
				break
			}
		} else if tok.Type != lexer.Modulo {
			break
		}
		op := p.consume().Value
		right, err := p.parseUnaryExpr()
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

func (p *Parser) parseUnaryExpr() (ast.Expr, *runtime.Error) {
	// Prefix operators: ++x, --x
	if p.peek().Type == lexer.Increment || p.peek().Type == lexer.Decrement {
		op := p.consume().Value
		operand, err := p.parseCallExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Operand: operand, Operator: op, Prefix: true}, nil
	}

	// Parse primary expression first
	expr, err := p.parseCallExpr()
	if err != nil {
		return nil, err
	}

	// Postfix operators: x++, x--
	if p.peek().Type == lexer.Increment || p.peek().Type == lexer.Decrement {
		op := p.consume().Value
		return &ast.UnaryExpr{Operand: expr, Operator: op, Prefix: false}, nil
	}

	return expr, nil
}

func (p *Parser) parseCallExpr() (ast.Expr, *runtime.Error) {
	callee, err := p.parseMemberExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == lexer.OpenParen {
		p.consume() // ( ->
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
		callee = &ast.CallExpr{Callee: callee, Args: args}
	}

	return callee, nil
}

func (p *Parser) parseMemberExpr() (ast.Expr, *runtime.Error) {
	obj, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.Dot {
		p.consume() // . ->
		prop, err := p.expect(lexer.Identifier, "Expected identifier after '.'")
		if err != nil {
			return nil, err
		}
		obj = &ast.MemberExpr{Object: obj, Property: &ast.Identifier{Symbol: prop.Value}}
	}
	return obj, nil
}

// parseprimary ->
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
	case lexer.Funct:
		return p.parseFunctionExpression()
	default:
		return nil, runtime.NewError(fmt.Sprintf("Unexpected token: %s", tok.Value), tok.Line, tok.Column)
	}
}

func (p *Parser) parseImportStatement() (ast.Stmt, *runtime.Error) {
	p.consume() // import ->
	strTok, err := p.expect(lexer.String, "Expected string path after 'import'")
	if err != nil {
		return nil, err
	}
	_, err = p.expect(lexer.As, "Expected 'as' after import path")
	if err != nil {
		return nil, err
	}
	aliasTok, err := p.expect(lexer.Identifier, "Expected identifier alias after 'as'")
	if err != nil {
		return nil, err
	}
	return &ast.ImportStatement{Path: strTok.Value, Alias: aliasTok.Value}, nil
}

func (p *Parser) parseFunctionDeclaration() (ast.Stmt, *runtime.Error) {
	p.consume() // funct ->
	nameTok, err := p.expect(lexer.Identifier, "Expected function name after 'funct'")
	if err != nil {
		return nil, err
	}
	_, err = p.expect(lexer.OpenParen, "Expected '(' after function name")
	if err != nil {
		return nil, err
	}
	params := []string{}
	if p.peek().Type != lexer.CloseParen {
		for {
			paramTok, err := p.expect(lexer.Identifier, "Expected parameter name")
			if err != nil {
				return nil, err
			}
			params = append(params, paramTok.Value)
			if p.peek().Type == lexer.CloseParen {
				break
			}
			_, err = p.expect(lexer.Comma, "Expected ',' or ')' in parameter list")
			if err != nil {
				return nil, err
			}
		}
	}
	_, err = p.expect(lexer.CloseParen, "Expected ')' after parameters")
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	return &ast.FunctionDeclaration{Name: nameTok.Value, Params: params, Body: body}, nil
}

// Parse function expression (anonymous function)
func (p *Parser) parseFunctionExpression() (ast.Expr, *runtime.Error) {
	// funct -> already consumed by parsePrimary
	_, err := p.expect(lexer.OpenParen, "Expected '(' after 'funct'")
	if err != nil {
		return nil, err
	}
	params := []string{}
	if p.peek().Type != lexer.CloseParen {
		for {
			paramTok, err := p.expect(lexer.Identifier, "Expected parameter name")
			if err != nil {
				return nil, err
			}
			params = append(params, paramTok.Value)
			if p.peek().Type == lexer.CloseParen {
				break
			}
			_, err = p.expect(lexer.Comma, "Expected ',' or ')' in parameter list")
			if err != nil {
				return nil, err
			}
		}
	}
	_, err = p.expect(lexer.CloseParen, "Expected ')' after parameters")
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	// Return a function declaration but as an expression
	return &ast.FunctionDeclaration{Name: "", Params: params, Body: body}, nil
}

func (p *Parser) parseReturnStatement() (ast.Stmt, *runtime.Error) {
	p.consume() // return ->
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStatement{Value: value}, nil
}

func (p *Parser) parseTryStatement() (ast.Stmt, *runtime.Error) {
	p.consume() // try
	tryBlock, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(lexer.Catch, "Expected 'catch' after try block")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(lexer.OpenParen, "Expected '(' after 'catch'")
	if err != nil {
		return nil, err
	}

	errorVar, err := p.expect(lexer.Identifier, "Expected error variable name in catch clause")
	if err != nil {
		return nil, err
	}

	_, err = p.expect(lexer.CloseParen, "Expected ')' after error variable")
	if err != nil {
		return nil, err
	}

	catchBlock, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	return &ast.TryStatement{
		TryBlock:   tryBlock,
		CatchBlock: catchBlock,
		ErrorVar:   errorVar.Value,
	}, nil
}

func (p *Parser) parseArrayLiteral() (ast.Expr, *runtime.Error) {
	elements := []ast.Expr{}
	// [ -> already consumed
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
	// { -> already consumed
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
