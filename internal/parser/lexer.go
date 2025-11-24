package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF

	// Identifiers and literals
	FIELD         // field name before colon
	STRING        // unquoted string
	NUMBER        // numeric value
	QUOTED_STRING // "quoted string"
	WILDCARD      // contains * or ?
	REGEX         // /pattern/

	// Operators and delimiters
	COLON    // :
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	LBRACE   // {
	RBRACE   // }
	PLUS     // +
	MINUS    // -
	CARET    // ^
	TILDE    // ~

	// Boolean operators
	AND // AND, &&
	OR  // OR, ||
	NOT // NOT, !

	// Range operators
	TO  // TO
	GT  // >
	GTE // >=
	LT  // <
	LTE // <=

	// Special queries
	EXISTS // _exists_
)

// Token represents a lexical token
type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

// String returns a string representation of the token type
func (tt TokenType) String() string {
	switch tt {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case FIELD:
		return "FIELD"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case QUOTED_STRING:
		return "QUOTED_STRING"
	case WILDCARD:
		return "WILDCARD"
	case REGEX:
		return "REGEX"
	case COLON:
		return "COLON"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case CARET:
		return "CARET"
	case TILDE:
		return "TILDE"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	case TO:
		return "TO"
	case GT:
		return "GT"
	case GTE:
		return "GTE"
	case LT:
		return "LT"
	case LTE:
		return "LTE"
	case EXISTS:
		return "EXISTS"
	default:
		return "UNKNOWN"
	}
}

// Lexer tokenizes OpenSearch query strings
type Lexer struct {
	input        string
	position     int  // current position in input
	readPosition int  // current reading position in input
	ch           byte // current char under examination
	line         int  // current line number (1-indexed)
	column       int  // current column number (1-indexed)
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar advances the position and reads the next character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

// peekChar returns the next character without advancing the position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekAhead returns the character n positions ahead without advancing
func (l *Lexer) peekAhead(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// currentPosition returns the current position
func (l *Lexer) currentPosition() Position {
	return Position{
		Offset: l.position,
		Line:   l.line,
		Column: l.column,
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Position = l.currentPosition()

	switch l.ch {
	case 0:
		tok.Type = EOF
		tok.Literal = ""
	case ':':
		tok.Type = COLON
		tok.Literal = string(l.ch)
		l.readChar()
	case '(':
		tok.Type = LPAREN
		tok.Literal = string(l.ch)
		l.readChar()
	case ')':
		tok.Type = RPAREN
		tok.Literal = string(l.ch)
		l.readChar()
	case '[':
		tok.Type = LBRACKET
		tok.Literal = string(l.ch)
		l.readChar()
	case ']':
		tok.Type = RBRACKET
		tok.Literal = string(l.ch)
		l.readChar()
	case '{':
		tok.Type = LBRACE
		tok.Literal = string(l.ch)
		l.readChar()
	case '}':
		tok.Type = RBRACE
		tok.Literal = string(l.ch)
		l.readChar()
	case '+':
		tok.Type = PLUS
		tok.Literal = string(l.ch)
		l.readChar()
	case '-':
		// Check for SQL comment indicator
		if l.peekChar() == '-' {
			tok.Type = ILLEGAL
			tok.Literal = "--"
			l.readChar()
			l.readChar()
			return tok
		}
		tok.Type = MINUS
		tok.Literal = string(l.ch)
		l.readChar()
	case '^':
		tok.Type = CARET
		tok.Literal = string(l.ch)
		l.readChar()
	case '~':
		tok.Type = TILDE
		tok.Literal = string(l.ch)
		l.readChar()
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = GTE
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = GT
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = LTE
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = LT
			tok.Literal = string(l.ch)
			l.readChar()
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok.Type = AND
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok.Type = OR
			tok.Literal = string(ch) + string(l.ch)
			l.readChar()
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
		}
	case '!':
		tok.Type = NOT
		tok.Literal = string(l.ch)
		l.readChar()
	case '"':
		tok.Type = QUOTED_STRING
		tok.Literal = l.readQuotedString()
	case '/':
		// Check for SQL comment or regex
		if l.peekChar() == '*' {
			tok.Type = ILLEGAL
			tok.Literal = "/*"
			return tok
		}
		// Try to read as regex
		if regex := l.tryReadRegex(); regex != "" {
			tok.Type = REGEX
			tok.Literal = regex
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
		}
	case ';':
		// Reject semicolons (SQL injection prevention)
		tok.Type = ILLEGAL
		tok.Literal = string(l.ch)
	default:
		if isLetter(l.ch) || l.ch == '_' || l.ch == '*' || l.ch == '?' {
			tok.Literal = l.readStringOrWildcard()
			tok.Type = l.lookupIdentType(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumberOrString()
			// Check if it's a pure number or mixed alphanumeric
			if containsLetters(tok.Literal) {
				tok.Type = STRING
			} else {
				tok.Type = NUMBER
			}
			return tok
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
		}
		l.readChar()
	}

	return tok
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier (field name or keyword)
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' ||
		(l.ch == '-' && l.peekChar() != '-') {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readStringOrWildcard reads a string that may contain wildcards
func (l *Lexer) readStringOrWildcard() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' || l.ch == '*' || l.ch == '?' ||
		(l.ch == '-' && l.peekChar() != '-') {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a numeric value
func (l *Lexer) readNumber() string {
	position := l.position
	hasDecimal := false

	for isDigit(l.ch) || (l.ch == '.' && !hasDecimal) {
		if l.ch == '.' {
			hasDecimal = true
		}
		l.readChar()
	}

	return l.input[position:l.position]
}

// readNumberOrString reads a value that starts with a digit but may contain letters
func (l *Lexer) readNumberOrString() string {
	position := l.position
	hasDecimal := false

	for isDigit(l.ch) || isLetter(l.ch) || (l.ch == '.' && !hasDecimal) || l.ch == '_' ||
		(l.ch == '-' && l.peekChar() != '-') {
		if l.ch == '.' {
			hasDecimal = true
		}
		l.readChar()
	}

	return l.input[position:l.position]
}

// readQuotedString reads a quoted string
func (l *Lexer) readQuotedString() string {
	var result strings.Builder
	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			if l.ch != 0 {
				// Handle escape sequence
				result.WriteByte(l.ch)
				l.readChar()
			}
		} else {
			result.WriteByte(l.ch)
			l.readChar()
		}
	}

	if l.ch == '"' {
		l.readChar() // skip closing quote
	}

	return result.String()
}

// readWildcard reads a wildcard pattern
func (l *Lexer) readWildcard() string {
	position := l.position

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '*' || l.ch == '?' || l.ch == '_' || l.ch == '-' {
		l.readChar()
	}

	return l.input[position:l.position]
}

// tryReadRegex attempts to read a regex pattern /pattern/
func (l *Lexer) tryReadRegex() string {
	if l.ch != '/' {
		return ""
	}

	position := l.position
	l.readChar() // skip opening /

	for l.ch != '/' && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			l.readChar() // skip escape
			if l.ch != 0 {
				l.readChar()
			}
		} else {
			l.readChar()
		}
	}

	if l.ch == '/' {
		l.readChar() // skip closing /
		// Return the pattern without the slashes
		return l.input[position+1 : l.position-1]
	}

	// Not a valid regex, reset
	l.position = position
	l.readPosition = position + 1
	l.ch = l.input[position]
	return ""
}

// lookupIdentType determines the token type for an identifier
func (l *Lexer) lookupIdentType(ident string) TokenType {
	switch ident {
	case "AND":
		return AND
	case "OR":
		return OR
	case "NOT":
		return NOT
	case "TO":
		return TO
	case "_exists_":
		return EXISTS
	default:
		// Check if it contains wildcards
		if strings.ContainsAny(ident, "*?") {
			return WILDCARD
		}
		return STRING
	}
}

// isLetter returns true if the character is a letter
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

// isDigit returns true if the character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// containsLetters returns true if the string contains any letters
func containsLetters(s string) bool {
	for _, ch := range s {
		if unicode.IsLetter(ch) {
			return true
		}
	}
	return false
}

// AllTokens returns all tokens from the input (useful for testing)
func (l *Lexer) AllTokens() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF || tok.Type == ILLEGAL {
			break
		}
	}
	return tokens
}

// TokenString returns a string representation of a token
func TokenString(tok Token) string {
	return fmt.Sprintf("%s(%q)", tok.Type, tok.Literal)
}
