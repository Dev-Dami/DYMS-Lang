package lexer

import (
	"fmt"
	"os"
	"unicode"
)

type TokenType int

const (
	// Literal Types
	Number TokenType = iota
	Identifier
	String

	// Keywords
	Let
	Var
	Const
	If
	Else
	For
	While
	ForRange
	True
	False
	Import
	As

	// Grouping * Operators
	BinaryOperator
	Equals
	ComparisonOperator
	LogicalOperator
	OpenParen
	CloseParen
	OpenBrace
	CloseBrace
	OpenBracket
	CloseBracket
	Colon
	Comma
	Dot
)

// Stringer for TokenType
func (t TokenType) String() string {
	switch t {
	case Number:
		return "Number"
	case Identifier:
		return "Identifier"
	case String:
		return "String"
	case Let:
		return "Let"
	case Var:
		return "Var"
	case Const:
		return "Const"
	case If:
		return "If"
	case Else:
		return "Else"
	case For:
		return "For"
	case While:
		return "While"
	case ForRange:
		return "ForRange"
	case True:
		return "True"
	case False:
		return "False"
	case BinaryOperator:
		return "BinaryOperator"
	case Equals:
		return "Equals"
	case ComparisonOperator:
		return "ComparisonOperator"
	case LogicalOperator:
		return "LogicalOperator"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case OpenBrace:
		return "OpenBrace"
	case CloseBrace:
		return "CloseBrace"
	case OpenBracket:
		return "OpenBracket"
	case CloseBracket:
		return "CloseBracket"
	case Colon:
		return "Colon"
	case Comma:
		return "Comma"
	case Dot:
		return "Dot"
	case Import:
		return "Import"
	case As:
		return "As"
	default:
		return "Unknown"
	}
}

type Token struct {
	Value  string
	Type   TokenType
	Line   int
	Column int
}

// create new token
func token(value string, t TokenType, line, column int) Token {
	return Token{Value: value, Type: t, Line: line, Column: column}
}

// keyword lookup
var keywords = map[string]TokenType{
	"let":       Let,
	"var":       Var,
	"const":     Const,
	"if":        If,
	"else":      Else,
	"for":       For,
	"while":     While,
	"for range": ForRange,
	"true":      True,
	"false":     False,
	"import":    Import,
	"as":        As,
}

func isAlpha(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isSkippable(ch rune) bool {
	// Treat Windows CR (\r) as skippable to support CRLF line endings
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r'
}

func isInt(ch rune) bool {
	return unicode.IsDigit(ch)
}

// Tokenizer
func Tokenize(sourceCode string) []Token {
	var tokens []Token
	src := []rune(sourceCode)
	line := 1
	col := 1

	for len(src) > 0 {
		ch := src[0]

		// Single-char tokens
		if ch == '"' {
			src = src[1:] // consume "
			col++
			str := ""
			for len(src) > 0 && src[0] != '"' {
				str += string(src[0])
				src = src[1:]
				col++
			}
			src = src[1:] // consume "
			col++
			tokens = append(tokens, token(str, String, line, col-len(str)-2))
		} else if ch == '(' {
			tokens = append(tokens, token(string(ch), OpenParen, line, col))
			src = src[1:]
			col++
		} else if ch == ')' {
			tokens = append(tokens, token(string(ch), CloseParen, line, col))
			src = src[1:]
			col++
		} else if ch == '{' {
			tokens = append(tokens, token(string(ch), OpenBrace, line, col))
			src = src[1:]
			col++
		} else if ch == '}' {
			tokens = append(tokens, token(string(ch), CloseBrace, line, col))
			src = src[1:]
			col++
		} else if ch == '[' {
			tokens = append(tokens, token(string(ch), OpenBracket, line, col))
			src = src[1:]
			col++
		} else if ch == ']' {
			tokens = append(tokens, token(string(ch), CloseBracket, line, col))
			src = src[1:]
			col++
		} else if ch == ':' {
			tokens = append(tokens, token(string(ch), Colon, line, col))
			src = src[1:]
			col++
		} else if ch == ',' {
			tokens = append(tokens, token(string(ch), Comma, line, col))
			src = src[1:]
			col++
		} else if ch == '.' {
			tokens = append(tokens, token(string(ch), Dot, line, col))
			src = src[1:]
			col++
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			if ch == '/' && len(src) > 1 && src[1] == '/' {
				// Skip comment
				for len(src) > 0 && src[0] != '\n' {
					src = src[1:]
					col++
				}
			} else {
				tokens = append(tokens, token(string(ch), BinaryOperator, line, col))
				src = src[1:]
				col++
			}
		} else if ch == '=' {
			if len(src) > 1 && src[1] == '=' {
				tokens = append(tokens, token("==", ComparisonOperator, line, col))
				src = src[2:]
				col += 2
			} else {
				tokens = append(tokens, token(string(ch), Equals, line, col))
				src = src[1:]
				col++
			}
		} else if ch == '!' {
			if len(src) > 1 && src[1] == '=' {
				tokens = append(tokens, token("!=", ComparisonOperator, line, col))
				src = src[2:]
				col += 2
			}
		} else if ch == '<' {
			if len(src) > 1 && src[1] == '=' {
				tokens = append(tokens, token("<=", ComparisonOperator, line, col))
				src = src[2:]
				col += 2
			} else {
				tokens = append(tokens, token("<", ComparisonOperator, line, col))
				src = src[1:]
				col++
			}
		} else if ch == '>' {
			if len(src) > 1 && src[1] == '=' {
				tokens = append(tokens, token(">=", ComparisonOperator, line, col))
				src = src[2:]
				col += 2
			} else {
				tokens = append(tokens, token(">", ComparisonOperator, line, col))
				src = src[1:]
				col++
			}
		} else if ch == '&' {
			if len(src) > 1 && src[1] == '&' {
				tokens = append(tokens, token("&&", LogicalOperator, line, col))
				src = src[2:]
				col += 2
			}
		} else if ch == '|' {
			if len(src) > 1 && src[1] == '|' {
				tokens = append(tokens, token("||", LogicalOperator, line, col))
				src = src[2:]
				col += 2
			}
		} else {
			// Multi-character tokens
			if isInt(ch) {
				startCol := col
				num := ""
				for len(src) > 0 && isInt(src[0]) {
					num += string(src[0])
					src = src[1:]
					col++
				}
				tokens = append(tokens, token(num, Number, line, startCol))
			} else if isAlpha(ch) {
				startCol := col
				ident := ""
				for len(src) > 0 && (isAlpha(src[0]) || isInt(src[0])) {
					ident += string(src[0])
					src = src[1:]
					col++
				}

				if ident == "for" && len(src) > 0 && src[0] == ' ' {
					tempSrc := src[1:]
					tempCol := col + 1
					nextIdent := ""
					for len(tempSrc) > 0 && (isAlpha(tempSrc[0]) || isInt(tempSrc[0])) {
						nextIdent += string(tempSrc[0])
						tempSrc = tempSrc[1:]
						tempCol++
					}
					if nextIdent == "range" {
						tokens = append(tokens, token("for range", ForRange, line, startCol))
						src = tempSrc
						col = tempCol
						continue
					}
				}

				if t, ok := keywords[ident]; ok {
					tokens = append(tokens, token(ident, t, line, startCol))
				} else {
					tokens = append(tokens, token(ident, Identifier, line, startCol))
				}
			} else if isSkippable(ch) {
				if ch == '\n' {
					line++
					col = 1
				} else {
					col++
				}
				src = src[1:] // skip whitespace
			} else {
				fmt.Fprintf(os.Stderr, "Unrecognized character: %d (%q) at line %d, column %d\n", ch, ch, line, col)
				os.Exit(1)
			}
		}
	}

	return tokens
}