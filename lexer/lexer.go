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

	// Keywords
	Let

	// Grouping * Operators
	BinaryOperator
	Equals
	OpenParen
	CloseParen
)

// Stringer for TokenType
func (t TokenType) String() string {
	switch t {
	case Number:
		return "Number"
	case Identifier:
		return "Identifier"
	case Let:
		return "Let"
	case BinaryOperator:
		return "BinaryOperator"
	case Equals:
		return "Equals"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	default:
		return "Unknown"
	}
}


type Token struct {
	Value string
	Type  TokenType
}

// create new token
func token(value string, t TokenType) Token {
	return Token{Value: value, Type: t}
}

// keyword lookup
var keywords = map[string]TokenType{
	"let": Let,
}

func isAlpha(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isSkippable(ch rune) bool {
	return ch == ' ' || ch == '\n' || ch == '\t'
}

func isInt(ch rune) bool {
	return unicode.IsDigit(ch)
}

// Tokenizer
func Tokenize(sourceCode string) []Token {
	var tokens []Token
	src := []rune(sourceCode)

	for len(src) > 0 {
		ch := src[0]

		// Single-char tokens
		if ch == '(' {
			tokens = append(tokens, token(string(ch), OpenParen))
			src = src[1:]
		} else if ch == ')' {
			tokens = append(tokens, token(string(ch), CloseParen))
			src = src[1:]
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			tokens = append(tokens, token(string(ch), BinaryOperator))
			src = src[1:]
		} else if ch == '=' {
			tokens = append(tokens, token(string(ch), Equals))
			src = src[1:]
		} else {
			// Multi-character tokens
			if isInt(ch) {
				num := ""
				for len(src) > 0 && isInt(src[0]) {
					num += string(src[0])
					src = src[1:]
				}
				tokens = append(tokens, token(num, Number))
			} else if isAlpha(ch) {
				ident := ""
				for len(src) > 0 && isAlpha(src[0]) {
					ident += string(src[0])
					src = src[1:]
				}

				if t, ok := keywords[ident]; ok {
					tokens = append(tokens, token(ident, t))
				} else {
					tokens = append(tokens, token(ident, Identifier))
				}
			} else if isSkippable(ch) {
				src = src[1:] // skip whitespace
			} else {
				fmt.Fprintf(os.Stderr, "Unrecognized character: %d (%q)\n", ch, ch)
				os.Exit(1)
			}
		}
	}

	return tokens
}
