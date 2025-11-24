package parser

import "fmt"

// Parser parses OpenSearch query strings into AST
type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
	errors  *ParseErrors
}

// NewParser creates a new parser for the given input
func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	p := &Parser{
		lexer:  lexer,
		errors: &ParseErrors{},
	}
	// Read two tokens to initialize current and peek
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

// Parse parses the query and returns the root AST node
func (p *Parser) Parse() (Node, error) {
	if p.current.Type == EOF {
		return nil, nil
	}

	expr := p.parseExpression(LOWEST)

	if p.errors.HasErrors() {
		return expr, p.errors
	}

	return expr, nil
}

// Operator precedence (lowest to highest)
const (
	LOWEST int = iota
	OR_PREC         // OR, ||
	AND_PREC        // AND, &&
	NOT_PREC        // NOT, !
	REQUIRED_PREC   // +term, -term
	FIELD_PREC      // field:value
	BOOST_PREC      // ^
	PREFIX          // highest
)

// precedence returns the precedence of the current token
func (p *Parser) currentPrecedence() int {
	switch p.current.Type {
	case OR:
		return OR_PREC
	case AND:
		return AND_PREC
	case NOT:
		return NOT_PREC
	case PLUS, MINUS:
		return REQUIRED_PREC
	case COLON:
		return FIELD_PREC
	case CARET:
		return BOOST_PREC
	default:
		return LOWEST
	}
}

// peekPrecedence returns the precedence of the peek token
func (p *Parser) peekPrecedence() int {
	switch p.peek.Type {
	case OR:
		return OR_PREC
	case AND:
		return AND_PREC
	case NOT:
		return NOT_PREC
	case PLUS, MINUS:
		return REQUIRED_PREC
	case COLON:
		return FIELD_PREC
	case CARET:
		return BOOST_PREC
	default:
		return LOWEST
	}
}

// parseExpression is the main recursive descent parser
func (p *Parser) parseExpression(precedence int) Node {
	// Prefix parsing
	var left Node

	switch p.current.Type {
	case LPAREN:
		left = p.parseGroupExpression()
	case NOT:
		left = p.parseNotExpression()
	case PLUS:
		left = p.parseRequiredExpression()
	case MINUS:
		left = p.parseProhibitedExpression()
	case LBRACKET, LBRACE:
		left = p.parseRangeExpression()
	case EXISTS:
		left = p.parseExistsQuery()
	case STRING, WILDCARD, NUMBER:
		left = p.parsePrimaryExpression()
	case QUOTED_STRING:
		left = p.parsePhraseExpression()
	default:
		p.addError(fmt.Sprintf("unexpected token: %s", p.current.Type), p.current.Position)
		p.nextToken()
		return nil
	}

	// Infix parsing - current token points to the operator after parsing left side
	for p.current.Type != EOF && p.current.Type != RPAREN {
		switch p.current.Type {
		case AND, OR:
			if precedence >= p.currentPrecedence() {
				return left
			}
			left = p.parseBinaryExpression(left)
		case COLON:
			if precedence >= p.currentPrecedence() {
				return left
			}
			// Check if this could be a field query
			if term, ok := left.(*TermQuery); ok {
				p.nextToken() // consume ':'
				left = p.parseFieldQuery(term.Term, term.Pos)
			} else {
				return left
			}
		case CARET:
			if precedence >= p.currentPrecedence() {
				return left
			}
			left = p.parseBoostExpression(left)
		case TILDE:
			if precedence >= p.currentPrecedence() {
				return left
			}
			left = p.parseFuzzyOrProximityExpression(left)
		default:
			return left
		}
	}

	// Handle implicit OR with adjacent terms
	if precedence == LOWEST && p.current.Type != EOF && p.current.Type != RPAREN &&
		p.current.Type != AND && p.current.Type != OR &&
		(p.current.Type == STRING || p.current.Type == WILDCARD || p.current.Type == NUMBER ||
			p.current.Type == QUOTED_STRING || p.current.Type == LPAREN || p.current.Type == PLUS ||
			p.current.Type == MINUS || p.current.Type == NOT || p.current.Type == EXISTS) {
		right := p.parseExpression(OR_PREC)
		if right != nil {
			return &BinaryOp{
				Op:    "OR",
				Left:  left,
				Right: right,
				Pos:   left.Position(),
			}
		}
	}

	return left
}

// parsePrimaryExpression parses a term, wildcard, or number
func (p *Parser) parsePrimaryExpression() Node {
	pos := p.current.Position
	lit := p.current.Literal

	var node Node
	switch p.current.Type {
	case STRING:
		node = &TermQuery{Term: lit, Pos: pos}
	case WILDCARD:
		node = &WildcardQuery{Pattern: lit, Pos: pos}
	case NUMBER:
		node = &TermQuery{Term: lit, Pos: pos}
	}

	p.nextToken()
	return node
}

// parsePhraseExpression parses a quoted phrase
func (p *Parser) parsePhraseExpression() Node {
	pos := p.current.Position
	phrase := p.current.Literal
	p.nextToken()

	return &PhraseQuery{Phrase: phrase, Pos: pos}
}

// parseGroupExpression parses (expr)
func (p *Parser) parseGroupExpression() Node {
	pos := p.current.Position
	p.nextToken() // consume '('

	expr := p.parseExpression(LOWEST)

	if p.peek.Type != RPAREN {
		p.addError("expected ')'", p.peek.Position)
		return expr
	}

	p.nextToken() // move to ')'
	p.nextToken() // consume ')'

	return &GroupQuery{Query: expr, Pos: pos}
}

// parseNotExpression parses NOT expr
func (p *Parser) parseNotExpression() Node {
	pos := p.current.Position
	op := p.current.Literal
	p.nextToken()

	operand := p.parseExpression(NOT_PREC)

	return &UnaryOp{
		Op:      op,
		Operand: operand,
		Pos:     pos,
	}
}

// parseRequiredExpression parses +term
func (p *Parser) parseRequiredExpression() Node {
	pos := p.current.Position
	p.nextToken()

	query := p.parseExpression(PREFIX)

	return &RequiredQuery{
		Query: query,
		Pos:   pos,
	}
}

// parseProhibitedExpression parses -term
func (p *Parser) parseProhibitedExpression() Node {
	pos := p.current.Position
	p.nextToken()

	query := p.parseExpression(PREFIX)

	return &ProhibitedQuery{
		Query: query,
		Pos:   pos,
	}
}

// parseBinaryExpression parses left OP right
func (p *Parser) parseBinaryExpression(left Node) Node {
	pos := p.current.Position
	op := p.current.Literal
	precedence := p.currentPrecedence()

	p.nextToken()

	right := p.parseExpression(precedence)

	return &BinaryOp{
		Op:    op,
		Left:  left,
		Right: right,
		Pos:   pos,
	}
}

// parseFieldQuery parses field:value
func (p *Parser) parseFieldQuery(field string, pos Position) Node {
	// Check if next is a group: field:(a OR b)
	if p.current.Type == LPAREN {
		return p.parseFieldGroupQuery(field, pos)
	}

	// Check if it's a range query
	if p.current.Type == LBRACKET || p.current.Type == LBRACE {
		return p.parseFieldRangeQuery(field, pos)
	}

	// Check for comparison operators
	if p.current.Type == GT || p.current.Type == GTE ||
		p.current.Type == LT || p.current.Type == LTE {
		return p.parseComparisonQuery(field, pos)
	}

	// Regular field:value
	value := p.parseValue()

	node := &FieldQuery{
		Field: field,
		Value: value,
		Pos:   pos,
	}

	// Check for fuzzy ~N
	if p.current.Type == TILDE {
		p.nextToken()
		distance := 2 // default fuzzy distance
		if p.current.Type == NUMBER {
			fmt.Sscanf(p.current.Literal, "%d", &distance)
			p.nextToken()
		}
		return &FuzzyQuery{
			Field:    field,
			Term:     value.Value().(string),
			Distance: distance,
			Pos:      pos,
		}
	}

	// Check for boost ^N
	if p.current.Type == CARET {
		p.nextToken()
		boost := 1.0
		if p.current.Type == NUMBER {
			fmt.Sscanf(p.current.Literal, "%f", &boost)
			p.nextToken()
		}
		return &BoostQuery{
			Query: node,
			Boost: boost,
			Pos:   pos,
		}
	}

	return node
}

// parseFieldGroupQuery parses field:(a OR b)
func (p *Parser) parseFieldGroupQuery(field string, pos Position) Node {
	p.nextToken() // consume '('

	var queries []Node
	for p.current.Type != RPAREN && p.current.Type != EOF {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			queries = append(queries, expr)
		}

		if p.current.Type == RPAREN {
			break
		}
	}

	if p.current.Type != RPAREN {
		p.addError("expected ')'", p.current.Position)
		return nil
	}

	p.nextToken() // consume ')'

	return &FieldGroupQuery{
		Field:   field,
		Queries: queries,
		Pos:     pos,
	}
}

// parseFieldRangeQuery parses field:[start TO end] or field:{start TO end}
func (p *Parser) parseFieldRangeQuery(field string, pos Position) Node {
	inclusive := p.current.Type == LBRACKET
	p.nextToken() // consume '[' or '{'

	start := p.parseValue()

	if p.current.Type != TO {
		p.addError("expected TO in range query", p.current.Position)
		return nil
	}

	p.nextToken() // consume 'TO'

	end := p.parseValue()

	var endInclusive bool
	if p.current.Type == RBRACKET {
		endInclusive = true
	} else if p.current.Type == RBRACE {
		endInclusive = false
	} else {
		p.addError("expected ']' or '}'", p.current.Position)
		return nil
	}

	p.nextToken() // consume ']' or '}'

	return &RangeQuery{
		Field:          field,
		Start:          start,
		End:            end,
		InclusiveStart: inclusive,
		InclusiveEnd:   endInclusive,
		Pos:            pos,
	}
}

// parseComparisonQuery parses field>value, field>=value, etc.
func (p *Parser) parseComparisonQuery(field string, pos Position) Node {
	op := p.current.Type
	p.nextToken()

	value := p.parseValue()

	// Convert comparison to range query
	var start, end ValueNode
	var inclusiveStart, inclusiveEnd bool

	switch op {
	case GT:
		start = value
		end = &TermValue{Term: "*", Pos: pos}
		inclusiveStart = false
		inclusiveEnd = false
	case GTE:
		start = value
		end = &TermValue{Term: "*", Pos: pos}
		inclusiveStart = true
		inclusiveEnd = false
	case LT:
		start = &TermValue{Term: "*", Pos: pos}
		end = value
		inclusiveStart = false
		inclusiveEnd = false
	case LTE:
		start = &TermValue{Term: "*", Pos: pos}
		end = value
		inclusiveStart = false
		inclusiveEnd = true
	}

	return &RangeQuery{
		Field:          field,
		Start:          start,
		End:            end,
		InclusiveStart: inclusiveStart,
		InclusiveEnd:   inclusiveEnd,
		Pos:            pos,
	}
}

// parseValue parses a value (term, phrase, wildcard, regex, number)
func (p *Parser) parseValue() ValueNode {
	pos := p.current.Position
	var value ValueNode

	switch p.current.Type {
	case STRING:
		value = &TermValue{Term: p.current.Literal, Pos: pos}
	case QUOTED_STRING:
		value = &PhraseValue{Phrase: p.current.Literal, Pos: pos}
	case WILDCARD:
		value = &WildcardValue{Pattern: p.current.Literal, Pos: pos}
	case REGEX:
		value = &RegexValue{Pattern: p.current.Literal, Pos: pos}
	case NUMBER:
		value = &NumberValue{Number: p.current.Literal, Pos: pos}
	default:
		p.addError(fmt.Sprintf("unexpected value type: %s", p.current.Type), pos)
		value = &TermValue{Term: "", Pos: pos}
	}

	p.nextToken()
	return value
}

// parseRangeExpression parses standalone range [50 TO 500]
func (p *Parser) parseRangeExpression() Node {
	pos := p.current.Position
	inclusive := p.current.Type == LBRACKET
	p.nextToken() // consume '[' or '{'

	start := p.parseValue()

	if p.current.Type != TO {
		p.addError("expected TO in range expression", p.current.Position)
		return nil
	}

	p.nextToken() // consume 'TO'

	end := p.parseValue()

	var endInclusive bool
	if p.current.Type == RBRACKET {
		endInclusive = true
	} else if p.current.Type == RBRACE {
		endInclusive = false
	} else {
		p.addError("expected ']' or '}'", p.current.Position)
		return nil
	}

	p.nextToken() // consume ']' or '}'

	return &RangeQuery{
		Field:          "", // no field specified
		Start:          start,
		End:            end,
		InclusiveStart: inclusive,
		InclusiveEnd:   endInclusive,
		Pos:            pos,
	}
}

// parseBoostExpression parses expr^boost
func (p *Parser) parseBoostExpression(expr Node) Node {
	pos := p.current.Position
	p.nextToken()

	boost := 1.0
	if p.current.Type == NUMBER {
		fmt.Sscanf(p.current.Literal, "%f", &boost)
		p.nextToken()
	}

	return &BoostQuery{
		Query: expr,
		Boost: boost,
		Pos:   pos,
	}
}

// parseFuzzyOrProximityExpression parses term~distance or "phrase"~distance
func (p *Parser) parseFuzzyOrProximityExpression(expr Node) Node {
	pos := p.current.Position
	p.nextToken()

	distance := 2 // default
	if p.current.Type == NUMBER {
		fmt.Sscanf(p.current.Literal, "%d", &distance)
		p.nextToken()
	}

	// Check if expr is a phrase or term
	if phrase, ok := expr.(*PhraseQuery); ok {
		return &ProximityQuery{
			Phrase:   phrase.Phrase,
			Distance: distance,
			Pos:      pos,
		}
	}

	if term, ok := expr.(*TermQuery); ok {
		return &FuzzyQuery{
			Field:    "",
			Term:     term.Term,
			Distance: distance,
			Pos:      pos,
		}
	}

	return expr
}

// parseExistsQuery parses _exists_:field
func (p *Parser) parseExistsQuery() Node {
	pos := p.current.Position
	p.nextToken() // consume '_exists_'

	if p.current.Type != COLON {
		p.addError("expected ':' after _exists_", p.current.Position)
		return nil
	}

	p.nextToken() // consume ':'

	if p.current.Type != STRING {
		p.addError("expected field name after _exists_:", p.current.Position)
		return nil
	}

	field := p.current.Literal
	p.nextToken()

	return &ExistsQuery{
		Field: field,
		Pos:   pos,
	}
}

// addError adds a parse error
func (p *Parser) addError(message string, pos Position) {
	p.errors.Add(NewParseError(message, pos))
}

// precedenceOfType returns the precedence of a token type
func precedenceOfType(tt TokenType) int {
	switch tt {
	case OR:
		return OR_PREC
	case AND:
		return AND_PREC
	case NOT:
		return NOT_PREC
	case PLUS, MINUS:
		return REQUIRED_PREC
	case COLON:
		return FIELD_PREC
	case CARET:
		return BOOST_PREC
	case STRING, WILDCARD, NUMBER, QUOTED_STRING:
		return LOWEST
	default:
		return LOWEST
	}
}
