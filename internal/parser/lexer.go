package parser

import (
	"strings"
	"unicode"
)

// TokenType represents the type of a token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenField
	TokenColon
	TokenString
	TokenNumber
	TokenAnd
	TokenOr
	TokenNot
	TokenPlus
	TokenMinus
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenTo
)

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// Lexer performs lexical analysis on a query string.
type Lexer struct {
	input string
	pos   int
	ch    rune
}

// NewLexer creates a new lexer.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar reads the next character.
func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
}

// peekChar looks at the next character without advancing.
func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := Token{Pos: l.pos - 1}

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
	case ':':
		tok.Type = TokenColon
		tok.Value = string(l.ch)
		l.readChar()
	case '+':
		tok.Type = TokenPlus
		tok.Value = string(l.ch)
		l.readChar()
	case '-':
		tok.Type = TokenMinus
		tok.Value = string(l.ch)
		l.readChar()
	case '!':
		tok.Type = TokenNot
		tok.Value = string(l.ch)
		l.readChar()
	case '(':
		tok.Type = TokenLParen
		tok.Value = string(l.ch)
		l.readChar()
	case ')':
		tok.Type = TokenRParen
		tok.Value = string(l.ch)
		l.readChar()
	case '[':
		tok.Type = TokenLBracket
		tok.Value = string(l.ch)
		l.readChar()
	case ']':
		tok.Type = TokenRBracket
		tok.Value = string(l.ch)
		l.readChar()
	case '"':
		tok.Type = TokenString
		tok.Value = l.readQuotedString()
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok.Type = TokenAnd
			tok.Value = "&&"
			l.readChar()
		} else {
			tok.Type = TokenField
			tok.Value = string(l.ch)
			l.readChar()
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok.Type = TokenOr
			tok.Value = "||"
			l.readChar()
		} else {
			tok.Type = TokenField
			tok.Value = string(l.ch)
			l.readChar()
		}
	default:
		if isLetter(l.ch) {
			tok.Value = l.readIdentifier()
			// Check for keywords
			switch strings.ToUpper(tok.Value) {
			case "AND":
				tok.Type = TokenAnd
			case "OR":
				tok.Type = TokenOr
			case "NOT":
				tok.Type = TokenNot
			case "TO":
				tok.Type = TokenTo
			default:
				tok.Type = TokenField
			}
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TokenNumber
			tok.Value = l.readNumber()
			return tok
		} else {
			tok.Type = TokenField
			tok.Value = string(l.ch)
			l.readChar()
		}
	}

	return tok
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() string {
	startPos := l.pos - 1
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}
	return l.input[startPos : l.pos-1]
}

// readNumber reads a numeric value.
func (l *Lexer) readNumber() string {
	startPos := l.pos - 1
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[startPos : l.pos-1]
}

// readQuotedString reads a quoted string.
func (l *Lexer) readQuotedString() string {
	l.readChar() // skip opening quote
	startPos := l.pos - 1
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
		}
		l.readChar()
	}
	value := l.input[startPos : l.pos-1]
	if l.ch == '"' {
		l.readChar() // skip closing quote
	}
	return value
}

// isLetter checks if a character is a letter.
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

// isDigit checks if a character is a digit.
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
