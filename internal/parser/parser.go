package parser

import (
	"fmt"
	"github.com/infiniv/rsearch/internal/translator"
)

// Parser parses query strings into AST nodes.
type Parser struct {
	lexer   *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}

// NewParser creates a new parser.
func NewParser(input string) *Parser {
	p := &Parser{
		lexer:  NewLexer(input),
		errors: []string{},
	}
	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// Parse parses the input and returns an AST node.
func (p *Parser) Parse() (translator.Node, error) {
	node := p.parseExpression(LowestPrecedence)
	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}
	return node, nil
}

// Operator precedence levels
const (
	LowestPrecedence = iota
	OrPrecedence
	AndPrecedence
	NotPrecedence
	PrefixPrecedence
)

// precedence returns the precedence of the current token.
func (p *Parser) precedence() int {
	switch p.curToken.Type {
	case TokenOr:
		return OrPrecedence
	case TokenAnd:
		return AndPrecedence
	case TokenNot:
		return NotPrecedence
	default:
		return LowestPrecedence
	}
}

// parseExpression parses an expression with the given precedence.
func (p *Parser) parseExpression(precedence int) translator.Node {
	// Parse prefix expression (unary operators, parentheses, field queries)
	left := p.parsePrefixExpression()
	if left == nil {
		return nil
	}

	// Parse infix expressions (binary operators)
	for p.curToken.Type != TokenEOF && precedence < p.precedence() {
		left = p.parseInfixExpression(left)
		if left == nil {
			return nil
		}
	}

	return left
}

// parsePrefixExpression parses prefix expressions.
func (p *Parser) parsePrefixExpression() translator.Node {
	switch p.curToken.Type {
	case TokenNot, TokenPlus, TokenMinus:
		return p.parseUnaryOp()
	case TokenLParen:
		return p.parseGroupedExpression()
	case TokenField:
		return p.parseFieldQuery()
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token: %v", p.curToken.Value))
		return nil
	}
}

// parseUnaryOp parses unary operations (NOT, +, -).
func (p *Parser) parseUnaryOp() translator.Node {
	op := p.curToken.Value
	tokenType := p.curToken.Type

	// Convert token type to operator string
	switch tokenType {
	case TokenNot:
		op = "NOT"
	case TokenPlus:
		op = "+"
	case TokenMinus:
		op = "-"
	}

	p.nextToken()

	operand := p.parseExpression(PrefixPrecedence)
	if operand == nil {
		return nil
	}

	return &translator.UnaryOp{
		Op:      op,
		Operand: operand,
	}
}

// parseGroupedExpression parses expressions in parentheses.
func (p *Parser) parseGroupedExpression() translator.Node {
	p.nextToken() // skip '('

	exp := p.parseExpression(LowestPrecedence)
	if exp == nil {
		return nil
	}

	if p.curToken.Type != TokenRParen {
		p.errors = append(p.errors, "expected ')'")
		return nil
	}

	p.nextToken() // skip ')'
	return exp
}

// parseFieldQuery parses field:value or field:[start TO end] queries.
func (p *Parser) parseFieldQuery() translator.Node {
	field := p.curToken.Value

	if p.peekToken.Type != TokenColon {
		p.errors = append(p.errors, fmt.Sprintf("expected ':' after field '%s'", field))
		return nil
	}

	p.nextToken() // skip field
	p.nextToken() // skip ':'

	// Check for range query
	if p.curToken.Type == TokenLBracket {
		return p.parseRangeQuery(field)
	}

	// Parse simple field:value query
	value := p.curToken.Value
	p.nextToken()

	return &translator.FieldQuery{
		Field: field,
		Value: value,
	}
}

// parseRangeQuery parses range queries like field:[start TO end].
func (p *Parser) parseRangeQuery(field string) translator.Node {
	inclusiveStart := true
	inclusiveEnd := true

	// Check for exclusive start bracket '{'
	if p.curToken.Type == TokenLBracket {
		p.nextToken()
	} else {
		// Handle '{' for exclusive range (future enhancement)
		inclusiveStart = false
		p.nextToken()
	}

	// Parse start value
	start := p.curToken.Value
	p.nextToken()

	// Expect TO keyword
	if p.curToken.Type != TokenTo {
		p.errors = append(p.errors, "expected TO in range query")
		return nil
	}
	p.nextToken()

	// Parse end value
	end := p.curToken.Value
	p.nextToken()

	// Check for exclusive end bracket '}'
	if p.curToken.Type == TokenRBracket {
		p.nextToken()
	} else {
		// Handle '}' for exclusive range (future enhancement)
		inclusiveEnd = false
		p.nextToken()
	}

	return &translator.RangeQuery{
		Field:          field,
		Start:          start,
		End:            end,
		InclusiveStart: inclusiveStart,
		InclusiveEnd:   inclusiveEnd,
	}
}

// parseInfixExpression parses infix expressions (binary operators).
func (p *Parser) parseInfixExpression(left translator.Node) translator.Node {
	if p.curToken.Type != TokenAnd && p.curToken.Type != TokenOr {
		return left
	}

	op := "AND"
	if p.curToken.Type == TokenOr {
		op = "OR"
	}

	precedence := p.precedence()
	p.nextToken()

	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}

	return &translator.BinaryOp{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

// Errors returns any parse errors.
func (p *Parser) Errors() []string {
	return p.errors
}
